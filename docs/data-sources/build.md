# `kustomization_build` Data Source

Data source to `kustomize build` a Kustomization and return a set of `ids` and hash map of `manifests` by `id`.

-> This data source was previously named `kustomization`. The name has been changed to follow the Terraform conventions. The previous name is still supported.

## Example Usage

```hcl
data "kustomization_build" "test" {
  path = "test_kustomizations/basic/initial"
  kustomize_options {
    load_restrictor = "none"
    enable_helm = true
    helm_path = "/path/to/helm"
  }
}

```

## Argument Reference

- `path` - (Required) Path to a kustomization directory.

### `kustomize_options` - (optional)

#### Child attributes

- `load_restrictor` - setting this to `"none"` disables load restrictions
- `enable_helm` - setting this to `true` allows referencing helm charts in the kustomization.yaml
- `helm_path` - set this to the path of the `helm` binary (defaults to: `helmV3`)

## Attribute Reference

- `ids` - Set of Kustomize resource IDs.
- `ids_prio` - List of Kustomize resource IDs grouped into three sets.
  - `ids_prio[0]`: `Kind: Namespace` and `Kind: CustomResourceDefinition`
  - `ids_prio[1]`: All `Kind`s not in `ids_prio[0]` or `ids_prio[2]`
  - `ids_prio[2]`: `Kind: MutatingWebhookConfiguration` and `Kind: ValidatingWebhookConfiguration`
- `manifests` - Map of JSON encoded Kubernetes resource manifests by ID.
