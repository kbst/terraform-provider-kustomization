package kustomize

import (
	"fmt"
	"io/ioutil"

	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// KubeClient holds kubernetes configuration
type KubeClient struct {
	Client dynamic.Interface
	Mapper *restmapper.DeferredDiscoveryRESTMapper
}

func initializeClient(raw, path, kubecontext string, incluster bool) (*KubeClient, error) {
	var config *rest.Config
	var err error

	if raw != "" {
		config, err = getClientConfig([]byte(raw), kubecontext)
		if err != nil {
			return nil, fmt.Errorf("provider kustomization: kubeconfig_raw: %s", err)
		}
	}

	if raw == "" && path != "" {
		data, err := readKubeconfigFile(path)
		if err != nil {
			return nil, fmt.Errorf("provider kustomization: kubeconfig_path: %s", err)
		}

		config, err = getClientConfig(data, kubecontext)
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

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	return &KubeClient{client, mapper}, nil
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
