# `kustomization_resource` Resource

Resource to provision JSON encoded Kubernetes resource manifests as produced by the `kustomization` data source on a Kubernetes cluster. Uses client-go dynamic client and uses server side dry runs to determine the Terraform plan for changing a resource.

## Example Usage

```hcl
data "kustomization_build" "test" {
  path = "test_kustomizations/basic/initial"
}

resource "kustomization_resource" "test" {
  for_each = data.kustomization_build.test.ids

  manifest = data.kustomization_build.test.manifests[each.value]
}

```

## Argument Reference

- `manifest` - (Required) JSON encoded Kubernetes resource manifest.
