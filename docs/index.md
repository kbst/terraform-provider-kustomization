# Kustomize Provider

This provider aims to solve 3 common issues of applying a kustomization using kubectl by integrating Kustomize and Terraform.

1. Lack of feedback what changes will be applied.
1. Resources from a previous apply not in the current apply are not purged.
1. Immutable changes like e.g. changing a deployment's selector cause the apply to fail mid way.

To solve this the provider uses the Terraform state to show changes to each resource individually during plan as well as track resources in need of purging.

It also uses [server side dry runs](https://kubernetes.io/docs/reference/using-api/api-concepts/#dry-run) to validate changes to the desired state and translate this into a Terraform plan that will show if a resource will be updated in-place or requires a delete and recreate to apply the changes.

As such it can be useful both to replace kustomize/kubectl integrated into a Terraform configuration as a provisioner as well as standalone `kubectl diff/apply` steps in CI/CD. The provider was primarily developed for Kubestack, the [Terraform GitOps framework](https://www.kubestack.com/), but is supported and tested to be used standalone.

## Example Usage

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

## Argument Reference

- `kubeconfig_path` - (Optional) Path to a kubeconfig file. Defaults to `KUBECONFIG`, `KUBE_CONFIG` or `~/.kube/config`.
- `kubeconfig_raw` - (Optional) Raw kubeconfig file. If kubeconfig_raw is set, kubeconfig_path is ignored.
- `context` - (Optional) Context to use in kubeconfig with multiple contexts, if not specified the default context is to be used.

## Imports

To import existing Kubernetes resources into the Terraform state for above usage example, use a command like below and replace `apps_v1_Deployment|test-basic|test` accordingly. Please note the single quotes required for most shells.

```
terraform import 'kustomization_resource.test["apps_v1_Deployment|test-basic|test"]' 'apps_v1_Deployment|test-basic|test'
```
