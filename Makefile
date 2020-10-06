build:
	go build -o .terraform/plugins/registry.terraform.io/kbst/kustomization/1.0.0/linux_amd64/terraform-provider-kustomization

test:
	TF_ACC=1 go test -v ./kustomize
