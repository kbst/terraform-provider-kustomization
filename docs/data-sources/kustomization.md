# `kustomization` Data Source

Data source to `kustomize build` a kustomization and return a set of `ids` and hash map of `manifests` by `id`.

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

- `path` - (Required) Path to a kustomization directory.

## Attribute Reference

- `ids` - Set of Kubernetes resource IDs.
- `manifests` - JSON encoded Kubernetes resource manifests.
