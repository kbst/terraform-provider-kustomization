build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o terraform.d/plugins/registry.terraform.io/kbst/kustomization/1.0.0/linux_amd64/terraform-provider-kustomization
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o terraform.d/plugins/registry.terraform.io/kbst/kustomization/1.0.0/darwin_amd64/terraform-provider-kustomization

test:
	TF_ACC=1 go test -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -v ./kustomize
