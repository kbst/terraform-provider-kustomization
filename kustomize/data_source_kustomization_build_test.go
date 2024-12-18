package kustomize

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceKustomization_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKustomizationConfig_basic("test_kustomizations/basic/initial"),
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

func testAccDataSourceKustomizationConfig_basic(path string) string {
	return fmt.Sprintf(`
data "kustomization_build" "test" {
	path = "%s"
}
`, path)
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
				Config: testAccDataSourceKustomizationConfig_helm("test_kustomizations/helm/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.kustomization_build.test", "id"),
					resource.TestCheckResourceAttrSet("data.kustomization_build.test", "path"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "path", "test_kustomizations/helm/initial"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "ids.#", "5"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "ids_prio.#", "3"),
					resource.TestCheckResourceAttr("data.kustomization_build.test", "manifests.%", "5"),
					resource.TestCheckOutput("service", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\",\"namespace\":\"test-basic\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":80}],\"selector\":{\"app\":\"nginx\"},\"type\":\"ClusterIP\"}}"),
					resource.TestCheckOutput("deployment", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\",\"namespace\":\"test-basic\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"nginx\"}},\"template\":{\"metadata\":{\"labels\":{\"app\":\"nginx\"}},\"spec\":{\"containers\":[{\"image\":\"nginx:6.0.10\",\"name\":\"test-basic\"}]}}}}"),
				),
			},
		},
	})
}

func testAccDataSourceKustomizationConfig_helm(path string) string {
	return fmt.Sprintf(`
data "kustomization_build" "test" {
	path = "%s"

	kustomize_options {
		enable_helm = true
		helm_path = "helm"
	}
}

output "service" {
	value = data.kustomization_build.test.manifests["_/Service/test-basic/nginx"]
}

output "deployment" {
	value = data.kustomization_build.test.manifests["apps/Deployment/test-basic/nginx"]
}
`, path)
}
