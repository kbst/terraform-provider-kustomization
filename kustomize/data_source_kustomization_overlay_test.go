package kustomize

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

//
//
// Basic acceptance test
func TestDataSourceKustomizationOverlay_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.kustomization_overlay.test", "id"),

					// Kustomization attributes
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "common_annotations.%", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "common_labels.%", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "components.#", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "config_map_generator.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "namespace", "test-overlay-basic"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "resources.#", "0"),

					// Generated
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "manifests.%", "1"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_basic() string {
	return `
data "kustomization_overlay" "test" {
	common_annotations = {}

	common_labels = {}

	components = []

	config_map_generator {
		name = "test-cm"
		literals = []
	}

	namespace = "test-overlay-basic"

	resources = []
}
`
}

//
//
// Test common_annotations attr
func TestDataSourceKustomizationOverlay_commonAnnotations(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_commonAnnotations(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check", "{\"apiVersion\":\"v1\",\"kind\":\"Namespace\",\"metadata\":{\"annotations\":{\"test-annotation\":\"true\"},\"name\":\"test-basic\"}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_commonAnnotations() string {
	return `
data "kustomization_overlay" "test" {
	common_annotations = {
		test-annotation: true
	}

	resources = [
		"../test_kustomizations/basic/initial",
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["~G_v1_Namespace|~X|test-basic"]
}
`
}

//
//
// Test common_labels attr
func TestDataSourceKustomizationOverlay_commonLabels(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_commonLabels(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check", "{\"apiVersion\":\"v1\",\"kind\":\"Namespace\",\"metadata\":{\"labels\":{\"test-label\":\"true\"},\"name\":\"test-basic\"}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_commonLabels() string {
	return `
data "kustomization_overlay" "test" {
	common_labels = {
		test-label: true
	}

	resources = [
		"../test_kustomizations/basic/initial",
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["~G_v1_Namespace|~X|test-basic"]
}
`
}

//
//
// Test components attr
func TestDataSourceKustomizationOverlay_components(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_components(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check", "{\"apiVersion\":\"v1\",\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"from-component-6ct58987ht\"}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_components() string {
	return `
data "kustomization_overlay" "test" {
	components = [
		"../test_kustomizations/component",
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["~G_v1_ConfigMap|~X|from-component-6ct58987ht"]
}
`
}

//
//
// Test config_map_generator attr
func TestDataSourceKustomizationOverlay_configMapGenerator(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationConfigMapGeneratorConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_cm1", "{\"apiVersion\":\"v1\",\"data\":{\"KEY1\":\"VALUE1\",\"KEY2\":\"VALUE2\"},\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"test-configmap1-gkfb9fdgch\"}}"),
					resource.TestCheckOutput("check_cm2", "{\"apiVersion\":\"v1\",\"data\":{\"KEY1\":\"VALUE1\",\"KEY2\":\"VALUE2\"},\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"test-configmap2-gkfb9fdgch\"}}"),
				),
			},
		},
	})
}

func testKustomizationConfigMapGeneratorConfig() string {
	return `
data "kustomization_overlay" "test" {
	config_map_generator {
		name = "test-configmap1"
		literals = [
			"KEY1=VALUE1",
			"KEY2=VALUE2"
		]
	}

	config_map_generator {
		name = "test-configmap2"
		literals = [
			"KEY1=VALUE1",
			"KEY2=VALUE2"
		]
	}
}

output "check_cm1" {
	value = data.kustomization_overlay.test.manifests["~G_v1_ConfigMap|~X|test-configmap1-gkfb9fdgch"]
}

output "check_cm2" {
	value = data.kustomization_overlay.test.manifests["~G_v1_ConfigMap|~X|test-configmap2-gkfb9fdgch"]
}
`
}

//
//
// Test namespace attr
func TestDataSourceKustomizationOverlay_namespace(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_namespace(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check", "{\"apiVersion\":\"v1\",\"kind\":\"Namespace\",\"metadata\":{\"name\":\"test-overlay-namespace\"}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_namespace() string {
	return `
data "kustomization_overlay" "test" {
	namespace = "test-overlay-namespace"

	resources = [
		"../test_kustomizations/basic/initial",
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["~G_v1_Namespace|~X|test-overlay-namespace"]
}

`
}

//
//
// Test resources attr
func TestDataSourceKustomizationOverlay_resources(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_resources(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check", "{\"apiVersion\":\"v1\",\"kind\":\"Namespace\",\"metadata\":{\"name\":\"test-basic\"}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_resources() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"../test_kustomizations/basic/initial",
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["~G_v1_Namespace|~X|test-basic"]
}

`
}
