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

module "test_nginx_ingress" {
  source  = "kbst.xyz/catalog/nginx/kustomization"
  version = "0.49.2-kbst.0"

  configuration_base_key = "default"
  configuration = {
    default = {}
  }
}

module "test_cert_manager" {
  source  = "kbst.xyz/catalog/cert-manager/kustomization"
  version = "1.5.4-kbst.0"

  configuration_base_key = "default"
  configuration = {
    default = {}
  }
}

module "test_prometheus" {
  source  = "kbst.xyz/catalog/prometheus/kustomization"
  version = "0.51.1-kbst.0"

  configuration_base_key = "default"
  configuration = {
    default = {}
  }
}

module "test_tekton" {
  source  = "kbst.xyz/catalog/tektoncd/kustomization"
  version = "0.28.1-kbst.0"

  configuration_base_key = "default"
  configuration = {
    default = {}
  }
}
