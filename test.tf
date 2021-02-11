terraform {
  required_providers {
    kustomization = {
      source = "kbst/kustomization"
      # all test versions are placed as 1.0.0
      # in .terraform/plugins for tests
      version = ">= 1.0.0"
    }
  }
  required_version = ">= 0.13"
}

data "kustomization_build" "test" {
  path = "kustomize/test_kustomizations/basic/initial"
}

resource "kustomization_resource" "from_build" {
  for_each = data.kustomization_build.test.ids

  manifest = data.kustomization_build.test.manifests[each.value]
}

data "kustomization_overlay" "test" {
  namespace = "test-overlay"

  resources = [
    "kustomize/test_kustomizations/basic/initial"
  ]
}

resource "kustomization_resource" "from_overlay" {
  for_each = data.kustomization_overlay.test.ids

  manifest = data.kustomization_overlay.test.manifests[each.value]
}
