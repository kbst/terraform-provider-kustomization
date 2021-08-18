module github.com/kbst/terraform-provider-kustomize

go 1.12

require (
	github.com/hashicorp/terraform-plugin-sdk v1.16.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v0.22.0
	k8s.io/kubectl v0.22.0
	sigs.k8s.io/kustomize/api v0.8.11
	sigs.k8s.io/kustomize/kyaml v0.11.0
)
