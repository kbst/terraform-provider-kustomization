module "kind_zero" {
  source = "github.com/kbst/terraform-kubestack//kind/cluster?ref=v0.13.0-beta.0"

  configuration = var.clusters["kind_zero"]
}
