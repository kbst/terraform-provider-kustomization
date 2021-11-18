package kustomize

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceKustomization_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKustomizationConfig_basic("test_kustomizations/basic/initial", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.kustomization_build.test", "id"),
					resource.TestCheckResourceAttrSet("data.kustomization_build.test", "path"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "path", "test_kustomizations/basic/initial"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "ids.#", "4"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "ids_prio.#", "3"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "manifests.%", "4"),
				),
			},
		},
	})
}

func testAccDataSourceKustomizationConfig_basic(path string, legacy bool) string {
	return fmt.Sprintf(`

provider "kustomization" {
	legacy_id_format = %v
}

data "kustomization_build" "test" {
	path = "%s"
}
`, legacy, path)
}

func TestAccDataSourceKustomization_legacyName(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKustomizationConfig_legacyName("test_kustomizations/basic/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.kustomization.test", "id"),
					resource.TestCheckResourceAttrSet("data.kustomization.test", "path"),
					resource.TestCheckResourceAttr("data.kustomization.test", "path", "test_kustomizations/basic/initial"),
					resource.TestCheckResourceAttr("data.kustomization.test", "ids.#", "4"),
					resource.TestCheckResourceAttr("data.kustomization.test", "ids_prio.#", "3"),
					resource.TestCheckResourceAttr("data.kustomization.test", "manifests.%", "4"),
				),
			},
		},
	})
}

func testAccDataSourceKustomizationConfig_legacyName(path string) string {
	return fmt.Sprintf(`
data "kustomization" "test" {
	path = "%s"
}
`, path)
}

func TestAccDataSourceKustomization_helmChart(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKustomizationConfig_helm("test_kustomizations/helm/initial", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.kustomization_build.test", "id"),
					resource.TestCheckResourceAttrSet("data.kustomization_build.test", "path"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "path", "test_kustomizations/helm/initial"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "ids_prio.#", "3"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "manifests.%", "2"),
					resource.TestCheckOutput("service", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"labels\":{\"name\":\"redis\"},\"name\":\"redis\"},\"spec\":{\"ports\":[{\"port\":6379,\"targetPort\":6379}],\"selector\":{\"name\":\"redis\"}}}"),
					resource.TestCheckOutput("deployment", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"labels\":{\"name\":\"redis\",\"release\":\"RELEASE-NAME\"},\"name\":\"redis\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"name\":\"redis\"}},\"template\":{\"metadata\":{\"labels\":{\"name\":\"redis\"}},\"spec\":{\"containers\":[{\"image\":\"redis:6.0.10\",\"imagePullPolicy\":\"IfNotPresent\",\"name\":\"simple-redis\",\"ports\":[{\"containerPort\":6379,\"name\":\"redis-port\",\"protocol\":\"TCP\"}],\"resources\":{\"limits\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"}}}]}}}}"),
				),
			},
		},
	})
}

func testAccDataSourceKustomizationConfig_helm(path string, legacy bool) string {
	return fmt.Sprintf(`

provider "kustomization" {
	legacy_id_format = %v
}

data "kustomization_build" "test" {
	path = "%s"

	kustomize_options = {
		enable_helm = true
		helm_path = "helm"
	}
}

output "service" {
	value = data.kustomization_build.test.manifests["_/Service/_/redis"]
}

output "deployment" {
	value = data.kustomization_build.test.manifests["apps/Deployment/_/redis"]
}
`, legacy, path)
}
