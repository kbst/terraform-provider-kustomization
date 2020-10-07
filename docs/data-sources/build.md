# `kustomization_build` Data Source

Data source to `kustomize build` a kustomization and return a set of `ids` and hash map of `manifests` by `id`.

-> This data source was previously named `kustomization`. The name has been changed to follow the Terraform conventions. The previous name is still supported.

## Example Usage

```hcl
data "kustomization_build" "test" {
  path = "test_kustomizations/basic/initial"
}

```

## Argument Reference

- `path` - (Required) Path to a kustomization directory.

## Attribute Reference

- `ids` - Set of Kubernetes resource IDs.
- `manifests` - JSON encoded Kubernetes resource manifests.
