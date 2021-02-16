variable "common_annotations" {
  type        = map(string)
  description = "Annotations to add to all resources."
  default     = null
}

variable "common_labels" {
  type        = map(string)
  description = "Labels to add to all resources."
  default     = null
}

variable "components" {
  type        = list(string)
  description = "Paths to Kustomize components."
  default     = null
}

variable "config_map_generator" {
  type        = list(any)
  description = "ConfigMaps to generate."
  default     = []
}

variable "crds" {
  type        = list(string)
  description = "List of paths to OpenAPI schema files as expected by Kustomize."
  default     = null
}

variable "generators" {
  type        = list(string)
  description = "List of paths to Kustomize generators."
  default     = null
}

variable "generator_options" {
  type        = any
  description = "Global Kustomize generator options."
  default     = null
}

variable "images" {
  type        = list(any)
  description = "List of images blocks."
  default     = []
}

variable "name_prefix" {
  type        = string
  description = "Prefix to add to names."
  default     = null
}

variable "namespace" {
  type        = string
  description = "The namespace to use."
  default     = null
}

variable "name_suffix" {
  type        = string
  description = "Suffix to add to names."
  default     = null
}

variable "patches" {
  type        = list(any)
  description = "List of patches."
  default     = []
}

variable "replicas" {
  type        = list(any)
  description = "List of replicas."
  default     = []
}

variable "secret_generator" {
  type        = list(any)
  description = "List of secret_generators."
  default     = []
}

variable "transformers" {
  type        = list(string)
  description = "List of paths to Kustomize transformers."
  default     = null
}

variable "vars" {
  type        = list(any)
  description = "List of vars."
  default     = []
}
