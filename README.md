# Terraform Provider Kustomize

![Run Tests](https://github.com/kbst/terraform-provider-kustomize/workflows/Run%20Tests/badge.svg?branch=master&event=push)

This provider aims to solve 3 common issues of applying a kustomization using kubectl by integrating Kustomize and Terraform.

1. Lack of feedback what changes will be applied.
1. Resources from a previous apply not in the current apply are not purged.
1. Immutable changes like e.g. changing a deployment's selector cause the apply to fail mid way.

To solve this the provider uses the Terraform state to show changes to each resource individually during plan as well as track resources in need of purging.

It also uses [server side dry runs](https://kubernetes.io/docs/reference/using-api/api-concepts/#dry-run) to validate changes to the desired state and translate this into a Terraform plan that will show if a resource will be updated in-place or requires a delete and recreate to apply the changes.

As such it can be useful both to replace kustomize/kubectl integrated into a Terraform configuration as a provisioner as well as standalone `kubectl diff/apply` steps in CI/CD.

## Using the Provider

The Terraform provider for Kustomize is available from the [Terraform registry](https://registry.terraform.io/providers/kbst/kustomization/latest). Please refert to the [documentation](https://registry.terraform.io/providers/kbst/kustomization/latest/docs) for information on how to use the `kustomization_build` and `kustomization_overlay` data sources, or the `kustomization_resource` resource.

## Development Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x
- [Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

## Building and Developing the Provider

To work on the provider, you need go installed on your machine (version 1.13.x tested). The provider uses go mod to manage its dependencies, so GOPATH is not required.

To compile the provider, run `make build` as shown below. This will build the provider and put the provider binary in the `terraform.d/plugins/linux_amd64/` directory.

```sh
$ make build
```

In order to test the provider, you can simply run the acceptance tests using `make test`. You can set the `KUBECONFIG` environment variable to point the tests to a specific cluster or set the context of your current config accordingly. The tests create namespaces on the current context. [Kind](https://github.com/kubernetes-sigs/kind) or [Minikube](https://github.com/kubernetes/minikube) clusters work well for testing.

```sh
$ make test
```
