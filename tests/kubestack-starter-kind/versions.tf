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
