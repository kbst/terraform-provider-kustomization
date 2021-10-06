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
      version = "0.7.0"
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
- `legacy_id_format` - currently defaults to `true` for backward compability, will default to `false` in future releases and be removed later again

## Migrating resource IDs from legacy format to desired format

To allow the kustomization provider to manage API version upgrades, the version has been removed
from resource IDs. As this is a breaking change, we are moving to what should be a nicer format

Legacy format: `apps_v1_Deployment|test-ns/test-deploy` or `~G_/_v1_/_Namespace|test-ns/test-svc`

New format: `apps/Deployment/test-ns/test-deploy` or `_/Service/test-ns/test-svc`

The general form is `group/Kind/namespace/name` with `_` as a placeholder for empty values (e.g. `_/Namespace/_/test-namespace`)

To use the new format of resource IDs, set `legacy_id_format` to `false` in the provider configuration and then migrate the existing state to use the new ID format.
```
terraform state list | while read line; do
  newstate=$(echo $line | sed -e 's:.*\["\([^_]*\)_\([^_]\)*_\([^|]*\)\|\([^|]*\)\|\([^"]*\)"\]:\1/\3/\4/\5:' -e 's/~[GX]/_/g')
  terraform state mv $line $newstate
done
```

## Imports

### New format

Set `legacy_id_format` to `false` in the provider settings and then run, for example:

```
terraform import 'kustomization_resource.test["apps/Deployment/test-namespace/test-deployment"]' apps/Deployment/test-namespace/test-deployment
```

### Legacy ID format

With `legacy_id_format` set to `true` (currently the default), use a command like below and replace `apps_v1_Deployment|test-namespace|test-deployment` accordingly.

-> Please note the single quotes required for most shells.

```
terraform import 'kustomization_resource.test["apps_v1_Deployment|test-namespace|test-deployment"]' 'apps_v1_Deployment|test-namespace|test-deployment'
```
