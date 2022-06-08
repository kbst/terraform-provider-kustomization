data "kustomization_build" "test_for_resource_config" {
  path = "kustomize/test_kustomizations/basic_for_resource_config/initial"
}

resource "kustomization_resource" "from_build_for_resource_config" {
  for_each = data.kustomization_build.test_for_resource_config.ids

  manifest        = data.kustomization_build.test_for_resource_config.manifests[each.value]
  kubeconfig_path = "~/.kube/config"
}
