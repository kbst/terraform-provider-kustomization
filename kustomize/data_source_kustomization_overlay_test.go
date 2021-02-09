package kustomize

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceKustomizationOverlay_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKustomizationOverlayConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.kustomization_overlay.test", "id"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "namespace", "test-overlay"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "resources.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "resources.0", "../test_kustomizations/basic/initial"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids.#", "4"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "manifests.%", "4"),
				),
			},
		},
	})
}

func testAccDataSourceKustomizationOverlayConfig_basic() string {
	return `
data "kustomization_overlay" "test" {
	namespace = "test-overlay"

	resources = [
		"../test_kustomizations/basic/initial",
	]
}
`
}
