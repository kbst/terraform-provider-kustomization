package main

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mitchellh/go-homedir"
)

// Config ...
type Config struct {
	Client dynamic.Interface
}

const kubeconfigDefault = "~/.kube/config"

// Provider ...
func Provider() *schema.Provider {
	p := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"kustomization_resource": kustomizationResource(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"kustomization": dataSourceKustomization(),
		},

		Schema: map[string]*schema.Schema{
			"kubeconfig_path": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc(
					[]string{
						"KUBE_CONFIG",
						"KUBECONFIG",
					},
					kubeconfigDefault),
				Description: fmt.Sprintf("Path to a kubeconfig file. Defaults to '%s'.", kubeconfigDefault),
			},
			"kubeconfig_raw": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Raw kubeconfig file. If kubeconfig_raw is set,  kubeconfig_path is ignored.",
			},
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		var data []byte
		var config *rest.Config
		var err error

		raw := d.Get("kubeconfig_raw").(string)
		data = []byte(raw)

		// try to get a config from kubeconfig_raw
		config, err = clientcmd.RESTConfigFromKubeConfig(data)
		if err != nil {
			// if kubeconfig_raw did not work, try kubeconfig_path
			path := d.Get("kubeconfig_path").(string)
			data, _ = readKubeconfigFile(path)

			config, err = clientcmd.RESTConfigFromKubeConfig(data)
			if err != nil {
				// if neither worked we fall back to an empty default config
				config = &rest.Config{}
			}
		}

		client, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		return &Config{client}, nil
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
