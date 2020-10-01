build:
	go build -o terraform.d/plugins/linux_amd64/terraform-provider-kustomization

test:
	TF_ACC=1 go test -v ./kustomize
