package kustomize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"


)

// Config ...
type Config struct {
	Client dynamic.Interface
	Mapper *restmapper.DeferredDiscoveryRESTMapper
	Mutex  *sync.Mutex
	// whether legacy IDs are produced or not
	LegacyIDs bool
}

// Provider ...
func Provider() *schema.Provider {
	p := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"kustomization_resource": kustomizationResource(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			// legacy name of the data source
			"kustomization": dataSourceKustomization(),
			// new name for the data source
			"kustomization_build": dataSourceKustomization(),

			// define overlay from TF
			"kustomization_overlay": dataSourceKustomizationOverlay(),
		},

		Schema: map[string]*schema.Schema{
			"kubeconfig_path": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("KUBECONFIG_PATH", nil),
				ExactlyOneOf: []string{"kubeconfig_path", "kubeconfig_raw", "kubeconfig_incluster", "kubernetes_provider_compat"},
				Description:  "Path to a kubeconfig file. Can be set using KUBECONFIG_PATH env var",
			},
			"kubeconfig_raw": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"kubeconfig_path", "kubeconfig_raw", "kubeconfig_incluster", "kubernetes_provider_compat"},
				Description:  "Raw kube config. If kubeconfig_raw is set, KUBECONFIG_PATH is ignored.",
			},
			"kubeconfig_incluster": {
				Type:         schema.TypeBool,
				Optional:     true,
				ExactlyOneOf: []string{"kubeconfig_path", "kubeconfig_raw", "kubeconfig_incluster", "kubernetes_provider_compat"},
				Description:  "Set to true when running inside a kubernetes cluster. If kubeconfig_incluster is set, KUBECONFIG_PATH is ignored.",
			},
			"context": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBECONFIG_CONTEXT", nil),
				Description: "Context to use in kubeconfig with multiple contexts, if not specified the default context is to be used.",
			},
			"legacy_id_format": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Deprecated:  "legacy_id_format will be removed in a future version",
				Description: "If legacy_id_format is true, then resource IDs will look like group_version_kind|namespace|name. If legacy_id_format is false, then resource IDs will look like group/kind/namespace/name",
			},
			// Subset of the official kubernetes provider configuration https://github.com/hashicorp/terraform-provider-kubernetes/blob/main/kubernetes/provider.go
			"kubernetes_provider_compat": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ExactlyOneOf: []string{"kubeconfig_path", "kubeconfig_raw", "kubeconfig_incluster", "kubernetes_provider_compat"},
				Description: "Rudimentary compatibility with the official kubernetes_provider",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("KUBE_HOST", ""),
							Description: "The hostname (in form of URI) of Kubernetes master.",
						},
						"cluster_ca_certificate": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("KUBE_CLUSTER_CA_CERT_DATA", ""),
							Description: "PEM-encoded root certificates bundle for TLS authentication.",
						},
						"token": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("KUBE_TOKEN", ""),
							Description: "Token to authenticate an service account",
						},
					},
				},
			},
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		var config *rest.Config
		var err error
		var compatOverrides *clientcmd.ConfigOverrides = nil

		raw := d.Get("kubeconfig_raw").(string)
		path := d.Get("kubeconfig_path").(string)
		incluster := d.Get("kubeconfig_incluster").(bool)
		context := d.Get("context").(string)

		if v, ok := d.GetOk("kubernetes_provider_compat"); ok {
			if spec, ok := v.([]interface{})[0].(map[string]interface{}); ok {
				overrides := &clientcmd.ConfigOverrides{}
				overrides.AuthInfo.Token = spec["token"].(string)
				if v, ok := d.GetOk("token"); ok {
					overrides.AuthInfo.Token = v.(string)
				}
				if v, ok := d.GetOk("cluster_ca_certificate"); ok {
					overrides.ClusterInfo.CertificateAuthorityData = bytes.NewBufferString(v.(string)).Bytes()
				}
				if v, ok := d.GetOk("host"); ok {

					host, _, err := rest.DefaultServerURL(v.(string), "", apimachineryschema.GroupVersion{}, true)
					if err != nil {
						return nil, fmt.Errorf("Failed to parse host: %s", err)
					}
					overrides.ClusterInfo.Server = host.String()
				}
				compatOverrides = overrides
			} else {
				return nil, fmt.Errorf("Failed to parse kubernetes_provider_compat")
			}
		}


		if raw != "" {
			config, err = getClientConfig([]byte(raw), context)
			if err != nil {
				return nil, fmt.Errorf("provider kustomization: kubeconfig_raw: %s", err)
			}
		}

		if raw == "" && path != "" {
			data, err := readKubeconfigFile(path)
			if err != nil {
				return nil, fmt.Errorf("provider kustomization: kubeconfig_path: %s", err)
			}

			config, err = getClientConfig(data, context)
			if err != nil {
				return nil, fmt.Errorf("provider kustomization: kubeconfig_path: %s", err)
			}
		}

		if incluster {
			config, err = rest.InClusterConfig()
			if err != nil {
				return nil, fmt.Errorf("provider kustomization: couldn't load in cluster config: %s", err)
			}
		}

		if compatOverrides != nil {
			cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&clientcmd.ClientConfigLoadingRules{}, compatOverrides)
			config, err = cc.ClientConfig()
			if err != nil {
				return nil, fmt.Errorf("provider kustomization: failed to construct client based on kubernetes_provider_compat")
			}

		}

		// empty default config required to support
		// using a cluster resource or data source
		// that may not exist yet, to configure the provider
		if config == nil {
			config = &rest.Config{}
		}

		// Increase QPS and Burst rate limits
		config.QPS = 120
		config.Burst = 240

		client, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("provider kustomization: %s", err)
		}

		dc, err := discovery.NewDiscoveryClientForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("provider kustomization: %s", err)
		}

		mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

		// Mutex to prevent parallel Kustomizer runs
		// temp workaround for upstream bug
		// https://github.com/kubernetes-sigs/kustomize/issues/3659
		mu := &sync.Mutex{}

		legacyIDs := d.Get("legacy_id_format").(bool)

		return &Config{client, mapper, mu, legacyIDs}, nil
	}

	return p
}

func readKubeconfigFile(s string) ([]byte, error) {
	p, err := homedir.Expand(s)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getClientConfig(data []byte, context string) (*rest.Config, error) {
	if len(context) == 0 {
		return clientcmd.RESTConfigFromKubeConfig(data)
	}

	rawConfig, err := clientcmd.Load(data)
	if err != nil {
		return nil, err
	}

	var clientConfig = clientcmd.NewNonInteractiveClientConfig(
		*rawConfig,
		context,
		&clientcmd.ConfigOverrides{CurrentContext: context},
		nil)

	return clientConfig.ClientConfig()
}
