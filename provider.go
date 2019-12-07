package main

import (
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// Config ...
type Config struct {
	Client dynamic.Interface
}

// Provider ...
func Provider() *schema.Provider {
	p := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"kustomization_resource": kustomizationResource(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"kustomization": dataSourceKustomization(),
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		kubeconfigPath := os.Getenv("KUBECONFIG")
		data, err := ioutil.ReadFile(kubeconfigPath)
		if err != nil {
			return nil, err
		}

		kubeConfig, err := clientcmd.NewClientConfigFromBytes(data)
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
