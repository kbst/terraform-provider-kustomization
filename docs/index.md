# Kustomize Provider

This provider allows building existing kustomizations using the `kustomization_build` data source or defining
dynamic kustomizations in HCL using the `kustomization_overlay` data source and applying the resources from
either kustomization against a Kubernetes cluster using the `kustomization_resource` resource.

The provider is maintained as part of the [Terraform GitOps framework Kubestack](https://www.kubestack.com/).

Using this Kustomize provider and Terraform has three main benefits compared to applying a kustomization using `kubectl`:

1. Running `terraform plan` will show a diff of the changes to be applied.
1. Deleted resources from a previous configuration will be purged.
1. Changes to immutable fields will generate a destroy and re-create plan.

As such the provider can be useful to replace kustomize/kubectl integrated into a Terraform configuration as a provisioner or to replace standalone `kubectl diff/apply` steps in CI/CD.

## Example Usage

```hcl
terraform {
  required_providers {
    kustomization = {
      source  = "kbst/kustomization"
      version = "0.5.0"
    }
  }
}

provider "kustomization" {
  # one of kubeconfig_path, kubeconfig_raw or kubeconfig_incluster must be set

  # kubeconfig_path = "~/.kube/config"
  # can also be set using KUBECONFIG_PATH environment variable

  # kubeconfig_raw = data.template_file.kubeconfig.rendered
  # kubeconfig_raw = yamlencode(local.kubeconfig)

  # kubeconfig_incluster = true
}

```

## Argument Reference

- `kubeconfig_path` - Path to a kubeconfig file. Can be set using `KUBECONFIG_PATH` environment variable.
- `kubeconfig_raw` - Raw kubeconfig file. If `kubeconfig_raw` is set, `kubeconfig_path` is ignored.
- `kubeconfig_incluster` - Set to `true` when running inside a kubernetes cluster.
- `context` - (Optional) Context to use in kubeconfig with multiple contexts, if not specified the default context is used.

## Imports

To import existing Kubernetes resources into the Terraform state, use a command like below and replace `apps_v1_Deployment|test-basic|test` accordingly.

-> Please note the single quotes required for most shells.

```
terraform import 'kustomization_resource.test["apps_v1_Deployment|test-basic|test"]' 'apps_v1_Deployment|test-basic|test'
```
