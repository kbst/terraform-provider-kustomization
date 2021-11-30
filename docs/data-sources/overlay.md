# `kustomization_overlay` Data Source

Data source to define and build a dynamic Kustomize overlay based on values coming from Terraform. Returns a set of `ids` and hash map of `manifests` by `id`.

Use `kustomization_overlay` to define attributes you would set in a [Kustomization file](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/) in YAML format, but using Terraform (HCL) syntax. Below are examples for each of the supported attributes.

The difference between the `kustomization_build` and the `kustomization_overlay` data sources is that `kustomization_build` does not allow customizing the Kubernetes resources using values coming from Terraform. Using `kustomization_overlay`, you can use Terraform references to generate new Kubernetes resources or patch inherited ones.

## Example Usage

The example shows a variety of uses for the Kustomize attributes in combination with values coming from Terraform.

```hcl
locals {
  label_key = "example-label"
  label_value = true

  cm_key = "KEY1"
  cm_value = "VALUE1"
}

data "template_file" "example" {
  template = <<-EOT
    ---
    config:
      key1: value1
      key2: value2
    name: example
  EOT
}

data "kustomization_overlay" "example" {
  common_annotations = {
    example-annotation: true
  }

  common_labels = {
    (local.label_key) = local.label_value
  }

  config_map_generator {
    name = "example-configmap1"
    behavior = "create"
    literals = [
      "${local.cm_key}=${local.cm_value}",
      "filename.yaml=${data.template_file.example.rendered}"
    ]
  }

  resources = [
    "path/to/kustomization/to/inherit/from",
    "path/to/kubernetes/resource.yaml",
  ]
  kustomize_options = {
    load_restrictor = "none"
  }
}

```

## Argument Reference

### `common_annotations` - (optional)

Set [Kustomize commonAnnotations](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/commonannotations/) using `common_annotations` key/value pairs.

#### Example

```hcl
data "kustomization_overlay" "example" {
  common_annotations = {
    example-annotation = true
  }
}
```

### `common_labels` - (optional)

Set [Kustomize commonLabels](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/commonlabels/) using `common_labels` key/value pairs. Sets `labels` and immutable `labelSelectors`.

#### Example

```hcl
# example shows how keys and values can be references
locals {
  label_key   = "example-label"
  label_value = true
}

data "kustomization_overlay" "example" {
  common_labels = {
    (local.label_key) = local.label_value
  }
}
```

### `components` - (optional)

Add one or more paths to [Kustomize components](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/components/) to inherit from.

#### Example

```hcl
data "kustomization_overlay" "example" {
  components = [
    "path/to/component",
    "path/to/another/component"
  ]
}
```

### `config_map_generator` - (optional)

Define one or more [Kustomize configMapGenerators](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/configmapgenerator/) using `config_map_generator` blocks.

#### Child attributes

- `name` set name of the generated resource
- `namespace` set namespace of the generated resource
- `behavior` control inheritance behavior, one of `create`, `replace` or `merge`
- `envs` list of paths to files to include as key/value pairs
- `files` list of paths to files to include as files
- `literals` list of `key=value` formatted strings to set as key/value pairs
- `options` set [`generator_options`](#generator_options---optional) specific to this resource

#### Example

```hcl
locals {
  cm_key   = "KEY1"
  cm_value = "VALUE1"
}

data "template_file" "example" {
  template = <<-EOT
    ---
    config:
      key1: value1
      key2: value2
    name: example
  EOT
}

data "kustomization_overlay" "example" {
  # a first ConfigMap
  config_map_generator {
    name     = "example-configmap1"
    behavior = "create"
    literals = [
      "${local.cm_key}=${local.cm_value}",
      "filename.yaml=${data.template_file.example.rendered}"
    ]

    options {
      disable_name_suffix_hash = true
    }
  }

  # a second ConfigMap
  config_map_generator {
    name = "example-configmap2"
    literals = [
      "KEY1=VALUE1",
      "KEY2=VALUE2"
    ]
    envs = [
      "path/to/properties.env"
    ]
    files = [
      "path/to/config/file.cfg"
    ]
  }
}
```

### `crds` - (optional)

One or more paths to CRD schema definitions as expected by Kustomize.

#### Example

```hcl
data "kustomization_overlay" "example" {
  crds = [
    "path/to/crd.yaml",
  ]
}
```

### `generators` - (optional)

One or more paths to Kustomize generators.

#### Example

```hcl
data "kustomization_overlay" "example" {
  generators = [
    "path/to/generator.yaml",
  ]
}
```

### `generator_options` - (optional)

Set options for all generators in this Kustomization.

#### Child attributes

- `labels` labels to add to generated resources
- `annotations` annotations to add to generated resources
- `disable_name_suffix_hash` whether to add hash suffix to resource name

#### Example

```hcl
data "kustomization_overlay" "example" {
  generator_options {
    labels = {
      example-label = "example"
    }

    annotations = {
      example-annotation = "example"
    }

    disable_name_suffix_hash = true
  }
}
```

### `images` - (optional)

Customize container images using `images` blocks.

#### Child attributes

- `name` image name
- `new_name` new image name
- `new_tag` new image tag
- `digest` image digest

#### Example

```hcl
data "kustomization_overlay" "example" {
  resources = [
    "path/to/another/kustomization",
  ]

  images {
    name = "oldname1"
    new_name = "newname1"
    new_tag = "newtag1"
    digest = "sha256:abcdefghijklmnop123456"
  }

  images {
    name = "oldname2"
    new_name = "newname2"
    new_tag = "newtag2"
    digest = "sha256:abcdefghijklmnop123457"
  }
}
```

### `kustomize_options` - (optional)

#### Child attributes

- `load_restrictor` - setting this to `"none"` disables load restrictions
- `enable_helm` - setting this to `true` enabled helm templating support
- `helm_path` - set this to the path to find the helm executable (default: `"helmV3"`)

#### Example

```hcl
data "kustomization_overlay" "example" {
  kustomize_options = {
    load_restrictor = "none"
    enable_helm = true
    helm_path = "/path/to/helm"
  }
}
```

### `name_prefix` - (optional)

Set a prefix to add to all resource names.

#### Example

```hcl
data "kustomization_overlay" "example" {
  name_prefix = "example-"
}
```

### `namespace` - (optional)

Set a namespace for all namespaced resources.

#### Example

```hcl
data "kustomization_overlay" "example" {
  namespace = "new-namespace"
}
```

### `name_suffix` - (optional)

Set a suffix to add to all resource names.

#### Example

```hcl
data "kustomization_overlay" "example" {
  name_suffix = "-example"
}
```

### `patches` - (optional)

Define [Kustomize patches](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/patches/) to modify Kubernetes resources using `patches` blocks.

#### Child attributes

- `path` path to a patch file on disk
- `patch` patch defined as an inline string
- `target` patch target, specified by: `group`, `version`, `kind`, `name`, `namespace`, `label_selector`, `annotation_selector`
- `options` - set `allow_kind_change` and/or `allow_name_change` to `true` to allow `kind` or `metadata.name` to be changed by the patch
  (only relevant for strategic merge patches, JSON patches ignore this setting)

#### Example

```hcl
data "kustomization_overlay" "example" {
  resources = [
    "path/to/kustomization",
  ]

  patches {
    path = "path/to/patch.yaml"
    target = {
      label_selector = "app=example,env=${terraform.workspace}"
    }
  }

  patches {
    target = {
      kind = "Namespace"
      name = "test-ns"
    }
    patch = <<-EOF
      kind: Namespace
      metadata:
        name: new-ns
    EOF
    options {
      allow_name_change = true
    }
  }

  patches {
    patch = <<-EOF
      - op: replace
        path: /spec/rules/0/http/paths/0/path
        value: /newpath
    EOF
    target = {
      group = "networking.k8s.io"
      version = "v1beta1"
      kind = "Ingress"
      name = "example"
      namespace = "example-basic"
      annotation_selector = "nginx.ingress.kubernetes.io/rewrite-target"
    }
  }
}
```

### `replicas` - (optional)

Set the [Kustomize replicas](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/replicas/) to change the number of replicas of a resource.

#### Child attributes

- `name` name of the Kubernetes resource to change replicas of
- `count` number of desired replicas

#### Example

```hcl
data "kustomization_overlay" "example" {
  replicas {
    name = "example-deployment"
    count = 3
  }

  replicas {
    name = "example-statefulset"
    count = 5
  }
}
```

### `resources` - (optional)

List of [Kustomization resources](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/resource/) to inherit from or include.

#### Example

```hcl
data "kustomization_overlay" "example" {
  resources = [
    "path/to/kustomization/to/inherit/from",
    "path/to/kubernetes/resource.yaml",
  ]
}
```

### `secret_generator` - (optional)

Define one or more [Kustomize secretGenerators](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/secretgenerator/) using `secret_generator` blocks.

#### Child attributes

- `name` set name of the generated resource
- `namespace` set namespace of the generated resource
- `behavior` control inheritance behavior, one of `create`, `replace` or `merge`
- `type` set the type of the generated Kubernetes secret
- `envs` list of paths to files to include as key/value pairs
- `files` list of paths to files to include as files
- `literals` list of `key=value` formatted strings to set as key/value pairs
- `options` set [`generator_options`](#generator_options---optional) specific to this resource

#### Example

```hcl
resource "random_password" "password" {
  length = 16
  special = true
  override_special = "_%@"
}

data "kustomization_overlay" "example" {
  secret_generator {
    name = "example-secret1"
    type = "Opaque"
    literals = [
      "password=${random_password.password.result}",
    ]

    options {
      disable_name_suffix_hash = true
    }
  }

  secret_generator {
    name = "example-secret2"
    literals = [
      "KEY1=VALUE1",
      "KEY2=VALUE2"
    ]
    envs = [
      "path/to/properties.env"
    ]
    files = [
      "path/to/config/file.cfg"
    ]
  }
}
```

### `transformers` - (optional)

List of paths to Kustomization transformers.

#### Example

```hcl
data "kustomization_overlay" "example" {
  transformers = [
    "path/to/kustomization/transformer.yaml"
  ]
}
```

### `vars` - (optional)

Define [Kustomize vars](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/vars/) to substitute name references. E.g. the name of a generated secret including its hash suffix.

#### Child attributes

- `name` name of the var
- `obj_ref` reference to the Kubernetes resource as specified by `api_version`, `kind` and `name`.
- `field_ref` reference to the attribute of the Kubernetes resource specified by `field_path`.

#### Example

```hcl
data "kustomization_overlay" "example" {
  resources = [
    "path/to/kustomization",
  ]

  secret_generator {
    name = "example-secret1"
    type = "Opaque"
    literals = [
      "password=${random_password.password.result}",
    ]
  }

  vars {
    name = "SECRET_NAME"
    obj_ref = {
      api_version = "v1"
      kind = "Secret"
      name = "example-secret1"
    }
    field_ref = {
      field_path = "metadata.name"
    }
  }

  patches {
    patch = <<-EOF
      - op: add
        path: /spec/template/spec/containers/0/env
        value: [{"name": "SECRET_NAME", "value": "$(SECRET_NAME)"}]
    EOF
    target = {
      group = "apps"
      version = "v1"
      kind = "Deployment"
      name = "example"
    }
  }
}
```

## Attribute Reference

- `ids` - Set of Kustomize resource IDs.
- `ids_prio` - List of Kustomize resource IDs grouped into three sets.
  - `ids_prio[0]`: `Kind: Namespace` and `Kind: CustomResourceDefinition`
  - `ids_prio[1]`: All `Kind`s not in `ids_prio[0]` or `ids_prio[2]`
  - `ids_prio[2]`: `Kind: MutatingWebhookConfiguration` and `Kind: ValidatingWebhookConfiguration`
- `manifests` - Map of JSON encoded Kubernetes resource manifests by ID.
