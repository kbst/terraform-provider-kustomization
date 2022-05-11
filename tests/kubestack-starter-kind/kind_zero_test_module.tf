module "kind_zero_test_module" {
  providers = {
    kustomization = kustomization.kind_zero
  }

  source = "github.com/kbst/catalog//src/test?ref=test-v0.0.1-kbst.1"

  configuration = {
    apps = {
      variant = "overlay"

      name_prefix = "prefix-"
      name_suffix = "-suffix"

      namespace = "test-module"

      additional_resources = [
        "${path.root}/manifests/extra_configmap.yaml"
      ]

      common_annotations = {
        "test-annotation" = "test"
      }

      common_labels = {
        "test-label" = "test"
      }

      labels = [{
        pairs = {
          "test-label-only" = "test"
        }
      }]

      generator_options = {
        annotations = {
          annotation-generated = "test"
        }

        labels = {
          label-generated = "test"
        }

        disable_name_suffix_hash = false
      }

      config_map_generator = [{
        name      = "test-generated"
        namespace = "test-module"
        behavior  = "create"
        literals = [
          "KEY=VALUE"
        ]
      }]

      secret_generator = [{
        name      = "secret-readme"
        namespace = "test-module"
        behavior  = "create"
        type      = "generic"
        files = [
          "${path.root}/README.md"
        ]
      }]

      images = [{
        name     = "busybox"
        new_name = "busybox"
        new_tag  = "latest"
      }]

      patches = [{
        patch = <<-EOF
          apiVersion: apps/v1
          kind: Deployment
          metadata:
            name: test
          spec:
            template:
              spec:
                containers:
                - name: busybox
                  command:
                  - sleep
                  - "30"
        EOF

        target = {
          group   = "apps"
          version = "v1"
          kind    = "Deployment"
          name    = "test"
        }
      }]

      replicas = [{
        name  = "test"
        count = 2
      }]
    }
  }
}
