# Terraform Provider Kustomize

![Run Tests](https://github.com/kbst/terraform-provider-kustomize/workflows/Run%20Tests/badge.svg?branch=master&event=push)

This provider is maintained as part of the [Terraform GitOps framework Kubestack](https://www.kubestack.com/).

## Using the Provider

The Terraform provider for Kustomize is available from the [Terraform registry](https://registry.terraform.io/providers/kbst/kustomization/latest). Please refer to the [documentation](https://registry.terraform.io/providers/kbst/kustomization/latest/docs) for information on how to use the `kustomization_build` and `kustomization_overlay` data sources, or the `kustomization_resource` resource.

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
