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
