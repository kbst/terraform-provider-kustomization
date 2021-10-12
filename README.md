<p align="center">
 <img src="./assets/favicon.png" alt="Kubestack, The Open Source Gitops Framework" width="25%" height="25%" />
</p>

<h1 align="center">Terraform Provider Kustomize</h1>

<div align="center">

[![Status](https://img.shields.io/badge/status-active-success.svg)]()
![Run Tests](https://github.com/kbst/terraform-provider-kustomize/workflows/Run%20Tests/badge.svg?branch=master&event=push)
[![GitHub Issues](https://img.shields.io/github/issues/kbst/terraform-provider-kustomization.svg)](https://github.com/kbst/terraform-provider-kustomization/issues)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/kbst/terraform-provider-kustomization.svg)](https://github.com/kbst/terraform-provider-kustomization/pulls)

</div>

<div align="center">

![GitHub Repo stars](https://img.shields.io/github/stars/kbst/terraform-provider-kustomization?style=social)
![Twitter Follow](https://img.shields.io/twitter/follow/kubestack?style=social)

</div>


<h3 align="center"><a href="#Contributing">Join Our Contributors!</a></h3>

<div align="center">

<a href="https://github.com/kbst/terraform-provider-kustomization/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=kbst/terraform-provider-kustomization&max=36" />
</a>

</div>

## Introduction

The Terraform provider for Kustomize is available from the [Terraform registry](https://registry.terraform.io/providers/kbst/kustomization/latest).

Please refer to the [documentation](https://registry.terraform.io/providers/kbst/kustomization/latest/docs) for information on how to use the `kustomization_build` and `kustomization_overlay` data sources, or the `kustomization_resource` resource.

This is a standalone Terraform Provider, but is also used in the [Terraform GitOps framework Kubestack](https://www.kubestack.com/).


## Getting Help

**Community Help**  
If you have any questions while following the tutorial, join the [#kubestack](https://app.slack.com/client/T09NY5SBT/CMBCT7XRQ) channel on the Kubernetes community. To create an account request an [invitation](https://slack.k8s.io/).

**Professional Services**  
For organizations interested in accelerating their GitOps journey, [professional services](https://www.kubestack.com/lp/professional-services) are available.


## Contributing
Contributions to the Kubestack framework are welcome and encouraged. Before contributing, please read the [Contributing](./CONTRIBUTING.md) and [Code of Conduct](./CODE_OF_CONDUCT.md) Guidelines.

One super simple way to contribute to the success of this project is to give it a star.  

<div align="center">

![GitHub Repo stars](https://img.shields.io/github/stars/kbst/terraform-provider-kustomization?style=social)

</div>


## Development Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x
- [Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)


## Building and Developing the Provider

To work on the provider, you need go installed on your machine. The provider uses go mod to manage its dependencies, so GOPATH is not required.

To compile the provider, run `make build` as shown below. This will build the provider and put the provider binary in the `terraform.d/plugins/linux_amd64/` directory.

```sh
make build
```

In order to test the provider, run the acceptance tests using `make test`. You have to set the `KUBECONFIG_PATH` environment variable to point the tests to a valid config file. Each tests uses an individual namespaces. [Kind](https://github.com/kubernetes-sigs/kind) or [Minikube](https://github.com/kubernetes/minikube) clusters work well for testing.

```sh
make test
```

### Running the provider under debug

Running in debug mode is a four step process - launch the plugin in debug mode under delve, connect your IDE to delve, connect terraform to the plugin, and
then run terraform. Instructions here are for Visual Studio Code, configuring other IDEs is likely to be a similar process

* Ensure you have [`delve` installed](https://github.com/go-delve/delve/tree/master/Documentation/installation)
* Build the plugin locally using `go build ./...`
* From the terraform repository you wish to debug, run the following
  ```
  PLUGIN_PROTOCOL_VERSIONS=5 dlv exec --api-version=2 --listen=127.0.0.1:49188 --headless /path/to/terraform-provider-kustomization/terraform-provider-kustomize -- --debug
  ```
  If you run this outside of the terraform repository (e.g. in the plugin source directory), kustomize gets very confused when trying to find files
* In Visual Studio Code, add the [following configuration to launch.json](https://code.visualstudio.com/docs/editor/debugging#_launch-configurations)
  ```
    {
        "name": "Connect to server",
        "type": "go",
        "request": "attach",
        "mode": "remote",
        "remotePath": "${workspaceFolder}",
        "port": 49188,
        "host": "127.0.0.1",
        "apiVersion": 2,
        "env": {
            "KUBECONFIG_PATH": "${HOME}/.kube/config"
        },
        "dlvLoadConfig": {
            "followPointers": true,
            "maxVariableRecurse": 1,
            "maxStringLen": 512,
            "maxArrayValues": 64,
            "maxStructFields": -1
        }
    }
  ```
* Click the debug button and start `Connect to server`.
* From the terminal you started `dlv` from, you will see something like:
  ```
  Provider server started; to attach Terraform, set TF_REATTACH_PROVIDERS to the following:
  {"registry.terraform.io/kbst/kustomization":{"Protocol":"grpc","Pid":13366,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/7c/wdp684z11w1_6cj0d6lgl8hw0000gn/T/plugin2650218557"}}}
  ```
  Set this value using export `TF_REATTACH_PROVIDERS={"registry...}}}` in the terminal you want to run terraform from
* Set any breakpoints you wish to use
* Run terraform


### Running a specific test in debug mode

* Add the following configuration to .vscode/launch.json, updating the `TestFunctionNameGoesHere`
  to the name of the function you want to test:

  ```
  {
      "name": "Launch test function",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/kustomize",
      "env": {
          "TF_ACC": "1",
          "KUBECONFIG_PATH": "${env:HOME}/.kube/config"
      },
      "args": [
          "-test.v",
          "-test.run",
          "^TestFunctionNameGoesHere$",
      ],
  }
  ```
* Set any breakpoints you want to set
* Click the debug button and choose "Launch test function" in the dropdown

## Kubestack Repositories
* [kbst/terraform-kubestack](https://github.com/kbst/terraform-kubestack)  
    * Terraform GitOps Framework - Everything you need to build reliable automation for AKS, EKS and GKE Kubernetes clusters in one free and open-source framework.
* [kbst/kbst](https://github.com/kbst/kbst)  
    * Kubestack Framework CLI - All-in-one CLI to scaffold your Infrastructure as Code repository and deploy your entire platform stack locally for faster iteration.
* [kbst/terraform-provider-kustomization](https://github.com/kbst/terraform-provider-kustomization) (this repository)  
    * Kustomize Terraform Provider - A Kubestack maintained Terraform provider for Kustomize, available in the [Terraform registry](https://registry.terraform.io/providers/kbst/kustomization/latest).
* [kbst/catalog](https://github.com/kbst/catalog)  
    * Catalog of cluster services as Kustomize bases - Continuously tested and updated Kubernetes services, installed and customizable using native Terraform syntax.

