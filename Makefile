build:
	go build -o terraform.d/plugins/linux_amd64/terraform-provider-kustomization

test:
	TF_ACC=1 go test -v ./kustomize

RELEASE := $(shell git describe --tags)

release-binaries:
	GOOS=linux GOARCH=amd64 go build -o terraform.d/plugins/linux_amd64/terraform-provider-kustomization
	tar -caf terraform-provider-kustomization-linux-amd64_$(RELEASE).tgz terraform.d/plugins/linux_amd64/terraform-provider-kustomization
	
	GOOS=darwin GOARCH=amd64 go build -o terraform.d/plugins/darwin_amd64/terraform-provider-kustomization
	tar -caf terraform-provider-kustomization-darwin-amd64_$(RELEASE).tgz terraform.d/plugins/darwin_amd64/terraform-provider-kustomization

	GOOS=windows GOARCH=amd64 go build -o terraform.d/plugins/windows_amd64/terraform-provider-kustomization
	tar -caf terraform-provider-kustomization-windows-amd64_$(RELEASE).tgz terraform.d/plugins/windows_amd64/terraform-provider-kustomization