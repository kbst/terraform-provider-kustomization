# `kustomization` Data Source

Data source to `kustomize build` a kustomization and return a set of `ids` and hash map of `manifests` by `id`.

## Example Usage

!> Please note the difference between the local provider name `kustomization` and the registry source `source = "kbst/kustomize"`. This unconventional naming requires specifying the provider attribute for every resource. To resolve this, this will be the last release of this provider as `kbst/kustomize` all future versions will be released as `kbst/kustomization`.

```hcl
terraform {
  required_providers {
    kustomization = {
      source  = "kbst/kustomize"
      version = "v0.2.0-beta.3"
    }
  }
  required_version = ">= 0.12"
}

provider "kustomization" {}

data "kustomization" "test" {
  provider = kustomization

  path = "test_kustomizations/basic/initial"
}

```

## Argument Reference

- `path` - (Required) Path to a kustomization directory.

## Attribute Reference

- `ids` - Set of Kubernetes resource IDs.
- `manifests` - JSON encoded Kubernetes resource manifests.
