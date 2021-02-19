# `kustomization_resource` Resource

Resource to provision JSON encoded Kubernetes manifests as produced by the `kustomization_build` or `kustomization_overlay` data sources on a Kubernetes cluster. Uses client-go dynamic client and uses server side dry runs to determine the Terraform plan for changing a resource.

### Terraform Limitation

Terraform providers can not control the Terraform dependency graph. But namespaced Kubernetes resources require the namespace to be created first. And there are other examples, like CRDs. To work around Terraform's limitation, this provider retries creating Kubernetes resources that depend on another Kubernetes resource. This works well in many cases but has a potential race condition. It is possible, for various reasons, that the resource's dependency does not get created within the number of retries. Terraform will continue with other resources, but not retry the already failed ones. Applying the failed resources, requires a second Terraform run.

One possible workaround is to increase the number of parallel resource operations, using Terraform's `-parallelism=n` parameter. This increases the chance, that the resource the provider is waiting for gets created in time. But increased parallelism is not always an option.

A better approach is to instruct Terraform to handle the resources in the correct order, using an explicit `depends_on`. For this reason, both data sources additionally return `ids_prio`, three sets of IDs grouped by the order they should be applied in.

Below two examples show both the simplified usage with `for_each` and `ids`. As well as the example with an explicit `depends_on` and `for_each` based on what priority bracket the ID is in.

## Example Usage

### Simple Example

```hcl
data "kustomization_build" "test" {
  path = "test_kustomizations/basic/initial"
}

resource "kustomization_resource" "test" {
  for_each = data.kustomization_build.test.ids

  manifest = data.kustomization_build.test.manifests[each.value]
}

```

### Explicit `depends_on` Example

```hcl
data "kustomization_build" "test" {
  path = "kustomize/test_kustomizations/basic/initial"
}

# first loop through resources in ids_prio[0]
resource "kustomization_resource" "p0" {
  for_each = data.kustomization_build.test.ids_prio[0]

  manifest = data.kustomization_build.test.manifests[each.value]
}

# then loop through resources in ids_prio[1]
# and set an explicit depends_on on kustomization_resource.p0
resource "kustomization_resource" "p1" {
  for_each = data.kustomization_build.test.ids_prio[1]

  manifest = data.kustomization_build.test.manifests[each.value]

  depends_on = [kustomization_resource.p0]
}

# finally, loop through resources in ids_prio[2]
# and set an explicit depends_on on kustomization_resource.p1
resource "kustomization_resource" "p2" {
  for_each = data.kustomization_build.test.ids_prio[2]

  manifest = data.kustomization_build.test.manifests[each.value]

  depends_on = [kustomization_resource.p1]
}

```

## Argument Reference

- `manifest` - (Required) JSON encoded Kubernetes resource manifest.
