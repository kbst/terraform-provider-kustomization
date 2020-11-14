module github.com/kbst/terraform-provider-kustomize

go 1.12

require (
	github.com/hashicorp/terraform-plugin-sdk v1.4.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.18.12
	k8s.io/apimachinery v0.18.12
	k8s.io/client-go v0.18.12
	sigs.k8s.io/kustomize/api v0.6.5
)
