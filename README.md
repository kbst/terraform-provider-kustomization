<p align="center">
 <img src="./assets/favicon.png" alt="Kubestack, The Open Source Gitops Framework" width="25%" height="25%" />
</p>

<h1 align="center">Terraform Provider Kustomize</h1>
<h3 align="center">Terraform Provider for the Kubestack Gitops Framework</h3>

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

The Terraform provider for Kustomize is available from the [Terraform registry](https://registry.terraform.io/providers/kbst/kustomization/latest). Please refer to the [documentation](https://registry.terraform.io/providers/kbst/kustomization/latest/docs) for information on how to use the `kustomization_build` and `kustomization_overlay` data sources, or the `kustomization_resource` resource.

This provider is maintained as part of the [Terraform GitOps framework Kubestack](https://www.kubestack.com/).


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


## Getting Started with Kubestack

For the easiest way to get started, [visit the official Kubestack quickstart](https://www.kubestack.com/infrastructure/documentation/quickstart). This tutorial will help you get started with the Kubestack GitOps framework. It is divided into three steps.

1. Develop Locally
    * Scaffold your repository and tweak your config in a local development environment that simulates your actual cloud configuration using Kubernetes in Docker (KinD).
3. Provision Infrastructure
    * Set-up cloud prerequisites and bootstrap Kubestack's environment and clusters on your cloud provider for the first time.
4. Set-up Automation
    * Integrate CI/CD to automate changes following Kubestack's GitOps workflow.

See the [`tests`](./tests) directory for an example of how to extend this towards multi-cluster and/or multi-cloud.


## Getting Help

**Official Documentation**  
Refer to the [official documentation](https://www.kubestack.com/framework/documentation) for a deeper dive into how to use and configure Kubetack.

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


## Kubestack Repositories
* [kbst/terraform-kubestack](https://github.com/kbst/terraform-kubestack)  
    * Terraform GitOps Framework - Everything you need to build reliable automation for AKS, EKS and GKE Kubernetes clusters in one free and open-source framework.
* [kbst/kbst](https://github.com/kbst/kbst)  
    * Kubestack Framework CLI - All-in-one CLI to scaffold your Infrastructure as Code repository and deploy your entire platform stack locally for faster iteration.
* [kbst/terraform-provider-kustomization](https://github.com/kbst/terraform-provider-kustomization) (this repository)  
    * Kustomize Terraform Provider - A Kubestack maintained Terraform provider for Kustomize, available in the [Terraform registry](https://registry.terraform.io/providers/kbst/kustomization/latest).
* [kbst/catalog](https://github.com/kbst/catalog)  
    * Catalog of cluster services as Kustomize bases - Continuously tested and updated Kubernetes services, installed and customizable using native Terraform syntax.

