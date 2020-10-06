# `kustomization_resource` Resource

Resource to provision JSON encoded Kubernetes resource manifests as produced by the `kustomization` data source on a Kubernetes cluster. Uses client-go dynamic client and uses server side dry runs to determine the Terraform plan for changing a resource.

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

resource "kustomization_resource" "test" {
  provider = kustomization

  for_each = data.kustomization.test.ids

  manifest = data.kustomization.test.manifests[each.value]
}

```

## Argument Reference

- `manifest` - (Required) JSON encoded Kubernetes resource manifest.
