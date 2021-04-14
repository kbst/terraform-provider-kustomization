data "kustomization_overlay" "test" {
  common_annotations = var.common_annotations

  common_labels = var.common_labels

  components = var.components

  dynamic "config_map_generator" {
    for_each = var.config_map_generator
    iterator = i
    content {
      name      = lookup(i.value, "name", null)
      namespace = lookup(i.value, "namespace", null)
      behavior  = lookup(i.value, "behavior", null)
      envs      = lookup(i.value, "envs", null)
      files     = lookup(i.value, "files", null)
      literals  = lookup(i.value, "literals", null)
      options {
        labels                   = lookup(i.value, "options", null) != null ? lookup(i.value["options"], "labels", null) : null
        annotations              = lookup(i.value, "options", null) != null ? lookup(i.value["options"], "annotations", null) : null
        disable_name_suffix_hash = lookup(i.value, "options", null) != null ? lookup(i.value["options"], "disable_name_suffix_hash", false) : false
      }
    }
  }

  crds = var.crds

  generators = var.generators

  dynamic "generator_options" {
    for_each = var.generator_options != null ? [var.generator_options] : []
    iterator = i
    content {
      labels                   = lookup(i.value, "labels", null)
      annotations              = lookup(i.value, "annotations", null)
      disable_name_suffix_hash = lookup(i.value, "disable_name_suffix_hash", false)
    }
  }

  dynamic "images" {
    for_each = var.images
    iterator = i
    content {
      name     = lookup(i.value, "name", null)
      new_name = lookup(i.value, "new_name", null)
      new_tag  = lookup(i.value, "new_tag", null)
      digest   = lookup(i.value, "digest", null)
    }
  }

  name_prefix = var.name_prefix

  namespace = var.namespace

  name_suffix = var.name_suffix

  dynamic "patches" {
    for_each = var.patches
    iterator = i
    content {
      path   = lookup(i.value, "path", null)
      patch  = lookup(i.value, "patch", null)
      target = [lookup(i.value, "target", null) != null ? i.value["target"] : {}]
    }
  }

  dynamic "replicas" {
    for_each = var.replicas
    iterator = i
    content {
      name  = lookup(i.value, "name", null)
      count = lookup(i.value, "count", null)
    }
  }

  resources = ["test_kustomizations/basic/initial"]

  dynamic "secret_generator" {
    for_each = var.secret_generator
    iterator = i
    content {
      name      = lookup(i.value, "name", null)
      namespace = lookup(i.value, "namespace", null)
      behavior  = lookup(i.value, "behavior", null)
      type      = lookup(i.value, "type", null)
      envs      = lookup(i.value, "envs", null)
      files     = lookup(i.value, "files", null)
      literals  = lookup(i.value, "literals", null)
      options {
        labels                   = lookup(i.value, "options", null) != null ? lookup(i.value["options"], "labels", null) : null
        annotations              = lookup(i.value, "options", null) != null ? lookup(i.value["options"], "annotations", null) : null
        disable_name_suffix_hash = lookup(i.value, "options", null) != null ? lookup(i.value["options"], "disable_name_suffix_hash", false) : false
      }
    }
  }

  dynamic "vars" {
    for_each = var.vars
    iterator = i
    content {
      name = lookup(i.value, "name", null)
      obj_ref {
        api_version = lookup(i.value, "obj_ref", null) != null ? lookup(i.value["obj_ref"], "api_version", null) : null
        group       = lookup(i.value, "obj_ref", null) != null ? lookup(i.value["obj_ref"], "group", null) : null
        version     = lookup(i.value, "obj_ref", null) != null ? lookup(i.value["obj_ref"], "version", null) : null
        kind        = lookup(i.value, "obj_ref", null) != null ? lookup(i.value["obj_ref"], "kind", null) : null
        name        = lookup(i.value, "obj_ref", null) != null ? lookup(i.value["obj_ref"], "name", null) : null
        namespace   = lookup(i.value, "obj_ref", null) != null ? lookup(i.value["obj_ref"], "namespace", null) : null
      }
      field_ref {
        field_path = lookup(i.value, "field_ref", null) != null ? lookup(i.value["field_ref"], "field_path", null) : null
      }
    }
  }
}

output "kustomization" {
  value = data.kustomization_overlay.test
}
