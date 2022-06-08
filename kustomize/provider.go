package kustomize

import (
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Config ...
type Config struct {
	KubeClient

	Mutex *sync.Mutex
	// whether legacy IDs are produced or not
	LegacyIDs             bool
	GzipLastAppliedConfig bool
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
			"legacy_id_format": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Deprecated:  "legacy_id_format will be removed in a future version",
				Description: "If legacy_id_format is true, then resource IDs will look like group_version_kind|namespace|name. If legacy_id_format is false, then resource IDs will look like group/kind/namespace/name",
			},
			"gzip_last_applied_config": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When 'true' compress the lastAppliedConfig annotation for resources that otherwise would exceed K8s' max annotation size. All other resources use the regular uncompressed annotation. Set to 'false' to disable compression entirely.",
			},
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		kubeconfigRaw := d.Get("kubeconfig_raw").(string)
		kubeconfigPath := d.Get("kubeconfig_path").(string)
		incluster := d.Get("kubeconfig_incluster").(bool)
		kubeContext := d.Get("context").(string)

		k, err := initializeClient(kubeconfigRaw, kubeconfigPath, kubeContext, incluster)

		if k == nil || err != nil {
			return nil, fmt.Errorf("provider kustomization: %s", err)
		}
		mu := &sync.Mutex{}

		legacyIDs := d.Get("legacy_id_format").(bool)
		gzipLastAppliedConfig := d.Get("gzip_last_applied_config").(bool)

		return &Config{*k, mu, legacyIDs, gzipLastAppliedConfig}, nil
	}

	return p
}
