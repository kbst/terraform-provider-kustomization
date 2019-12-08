package main

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"k8s.io/client-go/dynamic"
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
			"kubeconfig": {
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
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		kubeconfigPath := d.Get("kubeconfig").(string)
		kubeConfig, err := readKubeconfig(kubeconfigPath)
		if err != nil {
			return nil, err
		}

		config, err := kubeConfig.ClientConfig()
		if err != nil {
			return nil, err
		}

		client, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		return &Config{client}, nil
	}

	return p
}

func readKubeconfig(s string) (clientcmd.ClientConfig, error) {
	p, err := homedir.Expand(s)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	kubeConfig, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		return nil, err
	}

	return kubeConfig, nil
}
