# `kustomization_resource` Resource

Resource to provision JSON encoded Kubernetes manifests as produced by the `kustomization_build` or `kustomization_overlay` data sources on a Kubernetes cluster. Uses client-go dynamic client and server side dry runs to determine the Terraform plan for changing a resource.

### Terraform Limitation

Terraform providers can not control the Terraform dependency graph. But namespaced Kubernetes resources require the namespace to be created first. And there are other examples, like CRDs. To work around Terraform's limitation, this provider retries creating Kubernetes resources that depend on another Kubernetes resource. This works well in many cases but has a potential race condition. It is possible, for various reasons, that the resource's dependency does not get created within the number of retries. Terraform will continue with other resources, but not retry the already failed ones. Applying the failed resources, requires a second Terraform run.

One possible workaround is to increase the number of parallel resource operations, using Terraform's `-parallelism=n` parameter. This increases the chance, that the resource the provider is waiting for gets created in time. But increased parallelism is not always an option.

A better approach is to instruct Terraform to handle the resources in the correct order, using an explicit `depends_on`. For this reason, both data sources additionally return `ids_prio`, three sets of IDs grouped by the order they should be applied in.

In addition to the inability of a provider to control the Terraform dependency graph, marking an attribute sensitive, to hide it from the Terraform plan output, is not possible conditionally in the provider. As a result, the `manifest` attribute can't be marked sensitive for Kubernetes secrets, but kept non-sensitive for all other resources to keep the ability to preview changes. As a result, marking the `manifest` attribute sensitive for Kubernetes secrets, and potentially other resources, has to be handled conditionally in Terraform code.

The explicit `depends_on` for correct ordering of resources, and the conditional `sensitive` to prevent leaking secret values to the Terraform plan output make using the provider rather verbose. To make this easier to use, a convenience module is available, which handles all this inside the module and allows setting the Kustomizations as module variables, that are then passed to the `kustomization_overlay` data source. 

Below are two examples, one using the convenience module, and another one showing the explicit `depends_on` and `for_each`, as well as the conditional `sensitive`.

## Example Usage

### Module Example

```hcl
module "example_custom_manifests" {
  source  = "kbst.xyz/catalog/custom-manifests/kustomization"
  version = "0.3.0"

  configuration_base_key = "default"  # must match workspace name
  configuration = {
    default = {
      namespace = "example-${terraform.workspace}"

      resources = [
        "${path.root}/manifests/example/namespace.yaml",
        "${path.root}/manifests/example/deployment.yaml",
        "${path.root}/manifests/example/service.yaml"
      ]

      common_labels = {
        "env" = terraform.workspace
      }
    }
  }
}
```

Complete documentation of the custom-manifests convenience module can be found in the [Kubestack framework documentation](https://www.kubestack.com/framework/documentation/cluster-service-modules#custom-manifests).

### Provider Example

Usage of the provider requires one of the data sources, which return IDs and manifests as JSON strings, and the `kustomization_resource` to loop over the IDs using `for_each`, explicit `depends_on` as well as conditional `sensitive` on the `manifest` attribute.

```hcl
data "kustomization_build" "test" {
  path = "kustomize/test_kustomizations/basic/initial"
}

# first loop through resources in ids_prio[0]
resource "kustomization_resource" "p0" {
  for_each = data.kustomization_build.test.ids_prio[0]

  manifest = (
    contains(["_/Secret"], regex("(?P<group_kind>.*/.*)/.*/.*", each.value)["group_kind"])
    ? sensitive(data.kustomization_build.test.manifests[each.value])
    : data.kustomization_build.test.manifests[each.value]
  )
}

# then loop through resources in ids_prio[1]
# and set an explicit depends_on on kustomization_resource.p0
resource "kustomization_resource" "p1" {
  for_each = data.kustomization_build.test.ids_prio[1]

  manifest = (
    contains(["_/Secret"], regex("(?P<group_kind>.*/.*)/.*/.*", each.value)["group_kind"])
    ? sensitive(data.kustomization_build.test.manifests[each.value])
    : data.kustomization_build.test.manifests[each.value]
  )

  depends_on = [kustomization_resource.p0]
}

# finally, loop through resources in ids_prio[2]
# and set an explicit depends_on on kustomization_resource.p1
resource "kustomization_resource" "p2" {
  for_each = data.kustomization_build.test.ids_prio[2]

  manifest = (
    contains(["_/Secret"], regex("(?P<group_kind>.*/.*)/.*/.*", each.value)["group_kind"])
    ? sensitive(data.kustomization_build.test.manifests[each.value])
    : data.kustomization_build.test.manifests[each.value]
  )

  depends_on = [kustomization_resource.p1]
}

```

## Argument Reference

- `manifest` - (Required) JSON encoded Kubernetes resource manifest.
