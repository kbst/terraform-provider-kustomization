package kustomize

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mitchellh/go-homedir"
)

// Config ...
type Config struct {
	Client                dynamic.Interface
	Mapper                *restmapper.DeferredDiscoveryRESTMapper
	Extractor             metav1ac.UnstructuredExtractor
	Mutex                 *sync.Mutex
	GzipLastAppliedConfig bool
	ServerSideApply       bool
	ServerSideApplyForce  bool
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
				ExactlyOneOf: []string{"kubeconfig_path", "kubeconfig_raw", "kubeconfig_incluster"},
				Description:  "Path to a kubeconfig file. Can be set using KUBECONFIG_PATH env var",
			},
			"kubeconfig_raw": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"kubeconfig_path", "kubeconfig_raw", "kubeconfig_incluster"},
				Description:  "Raw kube config. If kubeconfig_raw is set, KUBECONFIG_PATH is ignored.",
			},
			"kubeconfig_incluster": {
				Type:         schema.TypeBool,
				Optional:     true,
				ExactlyOneOf: []string{"kubeconfig_path", "kubeconfig_raw", "kubeconfig_incluster"},
				Description:  "Set to true when running inside a kubernetes cluster. If kubeconfig_incluster is set, KUBECONFIG_PATH is ignored.",
			},
			"context": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBECONFIG_CONTEXT", nil),
				Description: "Context to use in kubeconfig with multiple contexts, if not specified the default context is to be used.",
			},
			"gzip_last_applied_config": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When 'true' compress the lastAppliedConfig annotation for resources that otherwise would exceed K8s' max annotation size. All other resources use the regular uncompressed annotation. Set to 'false' to disable compression entirely.",
			},
			"server_side_apply": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When 'true' use server-side apply.",
			},
			"server_side_apply_force": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When 'true' the server-side apply will overwrite any managed fields that another owner has claimed.",
			},
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		var config *rest.Config
		var err error

		raw := d.Get("kubeconfig_raw").(string)
		path := d.Get("kubeconfig_path").(string)
		incluster := d.Get("kubeconfig_incluster").(bool)
		context := d.Get("context").(string)

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

		cDc := memory.NewMemCacheClient(dc)

		mapper := restmapper.NewDeferredDiscoveryRESTMapper(cDc)

		extractor, err := metav1ac.NewUnstructuredExtractor(cDc)
		if err != nil {
			return nil, fmt.Errorf("provider kustomization: %s", err)
		}

		// Mutex to prevent parallel Kustomizer runs
		// temp workaround for upstream bug
		// https://github.com/kubernetes-sigs/kustomize/issues/3659
		mu := &sync.Mutex{}

		gzipLastAppliedConfig := d.Get("gzip_last_applied_config").(bool)
		serverSideApply := d.Get("server_side_apply").(bool)
		serverSideApplyForce := d.Get("server_side_apply_force").(bool)

		return &Config{
			Client:                client,
			Mapper:                mapper,
			Extractor:             extractor,
			Mutex:                 mu,
			GzipLastAppliedConfig: gzipLastAppliedConfig,
			ServerSideApply:       serverSideApply,
			ServerSideApplyForce:  serverSideApplyForce,
		}, nil
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

	var clientConfig clientcmd.ClientConfig = clientcmd.NewNonInteractiveClientConfig(
		*rawConfig,
		context,
		&clientcmd.ConfigOverrides{CurrentContext: context},
		nil)

	return clientConfig.ClientConfig()
}
