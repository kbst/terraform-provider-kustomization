# Terraform Provider Kustomize

![Run Tests](https://github.com/kbst/terraform-provider-kustomize/workflows/Run%20Tests/badge.svg?branch=master&event=push)

This provider aims to solve 3 common issues of applying a kustomization using kubectl by integrating Kustomize and Terraform.

 1. Lack of feedback what changes will be applied.
 1. Resources from a previous apply not in the current apply are not purged.
 1. Immutable changes like e.g. changing a deployment's selector cause the apply to fail mid way.

To solve this the provider uses the Terraform state to show changes to each resource individually during plan as well as track resources in need of purging.

It also uses [server side dry runs](https://kubernetes.io/docs/reference/using-api/api-concepts/#dry-run) to validate changes to the desired state and translate this into a Terraform plan that will show if a resource will be updated in-place or requires a delete and recreate to apply the changes.

As such it can be useful both to replace kustomize/kubectl integrated into a Terraform configuration as a provisioner as well as standalone `kubectl diff/apply` steps in CI/CD.

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

## Usage

```hcl
data "kustomization" "example" {
  # path to kustomization directory
  path = "test_kustomizations/basic/initial"
}

resource "kustomization_resource" "example" {
  for_each = data.kustomization.example.ids

  manifest = data.kustomization.example.manifests[each.value]
}

```

## Configuring the provider

```hcl
provider "kustomization" {
  # optional path to kubeconfig file
  # falls back to KUBECONFIG or KUBE_CONFIG env var
  # or finally '~/.kube/config'
  kubeconfig_path = "/path/to/kubeconfig/file"

  # optional raw kubeconfig string
  # overwrites kubeconfig_path
  kubeconfig_raw = data.template_file.kubeconfig.rendered

  # optional context to use in kubeconfig with multiple contexts
  # if unspecified, the default (current) context is used
  context = "my-context"
}
```

## State import for kustomization_resource

To import existing Kubernetes resources into the Terraform state for above usage example, use a command like below and replace `apps_v1_Deployment|test-basic|test` accordingly. Please note the single quotes required for most shells.

```
terraform import 'kustomization_resource.test["apps_v1_Deployment|test-basic|test"]' 'apps_v1_Deployment|test-basic|test'
```

## Limitations
- To get rid of the noises with various "server-side managed" changing fields such as `managedFields`, `creationTimestamp`, `resourceVersion`, `status` etc. for a kubernetes resource in the terraform plan diff, we use a data source "targeted" strategy when doing the plan diff, i.e., only fields present in your kustomization code are involved in producing the terraform plan diff against the actual kubenetes resource in cluster.
- When doing an import, terraform is not aware of the kustomization data source. This is a known [limitation of terraform](https://www.terraform.io/docs/commands/import.html#provider-configuration). That way we have to import a minimal part of the resource into state (only identity of the resource). After importing, the first terraform apply could compare based on the targeted data source and therefore sync the state to the desired field values according to what you have in kustomization.

## Building and Developing the Provider

To work on the provider, you need go installed on your machine (version 1.13.x tested). The provider uses go mod to manage its dependencies, so GOPATH is not required.

To compile the provider, run `go build` as shown below. This will build the provider and put the provider binary in the `terraform.d/plugins/linux_amd64/` directory. The provider has not been tested yet on other platforms.

```sh
$ go build -o terraform.d/plugins/linux_amd64/terraform-provider-kustomization
```

In order to test the provider, you can simply run the acceptance tests using `go test`. You can set the `KUBECONFIG` environment variable to point the tests to a specific cluster or set the context of your current config accordingly. The tests create namespaces on the current context. [Kind](https://github.com/kubernetes-sigs/kind) or [Minikube](https://github.com/kubernetes/minikube) clusters work well for testing.

```sh
$ TF_ACC=1 go test -v ./kustomize
```
