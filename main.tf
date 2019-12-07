data "kustomization" "test" {
  path = "test_kustomizations/basic/initial"
}

resource "kustomization_resource" "test" {
  for_each = data.kustomization.test.ids

  manifest = data.kustomization.test.manifests[each.value]
}
