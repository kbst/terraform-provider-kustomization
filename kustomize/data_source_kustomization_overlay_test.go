package kustomize

import (
	"regexp"
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
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "crds.#", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "images.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "namespace", "test-overlay-basic"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "replicas.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "resources.#", "0"),

					// Generated
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids.#", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "manifests.%", "0"),
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

	config_map_generator {}

	crds = []

	images {}

	name_prefix = "test-"

	namespace = "test-overlay-basic"

	name_suffix = "-test"

	replicas {}

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
		"test_kustomizations/basic/initial",
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
		"test_kustomizations/basic/initial",
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
		"test_kustomizations/component",
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
					resource.TestCheckOutput("check_cm2", "{\"apiVersion\":\"v1\",\"data\":{\"ENV1\":\"VALUE1\",\"ENV2\":\"VALUE2\",\"KEY1\":\"VALUE1\",\"KEY2\":\"VALUE2\",\"properties.env\":\"ENV1=VALUE1\\nENV2=VALUE2\\n\"},\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"test-configmap2-5tgmgc9cmf\"}}"),
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
		behavior = "create"
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
		envs = [
			"test_kustomizations/_test_files/properties.env"
		]
		files = [
			"test_kustomizations/_test_files/properties.env"
		]
	}
}

output "check_cm1" {
	value = data.kustomization_overlay.test.manifests["~G_v1_ConfigMap|~X|test-configmap1-gkfb9fdgch"]
}

output "check_cm2" {
	value = data.kustomization_overlay.test.manifests["~G_v1_ConfigMap|~X|test-configmap2-5tgmgc9cmf"]
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
		"test_kustomizations/basic/initial",
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["~G_v1_Namespace|~X|test-overlay-namespace"]
}

`
}

//
//
// Test name_prefix and name_suffix attr
func TestDataSourceKustomizationOverlay_name_prefix_suffix(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_name_prefix_suffix(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test-test-test\",\"namespace\":\"test-basic\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":80}],\"selector\":{\"app\":\"test\"},\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_name_prefix_suffix() string {
	return `
data "kustomization_overlay" "test" {
	name_prefix = "test-"
	name_suffix = "-test"

	resources = [
		"test_kustomizations/basic/initial",
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["~G_v1_Service|test-basic|test-test-test"]
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
		"test_kustomizations/basic/initial",
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["~G_v1_Namespace|~X|test-basic"]
}

`
}

//
//
// Test crds attr
func TestDataSourceKustomizationOverlay_crds(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_crds(),
				// we only need to validate we pass the value to Kustomize
				// this test verifies that the path by providing an invalid OpenAPI spec
				// the Kustomize error proves the path was passed correctly
				ExpectError: regexp.MustCompile("json: cannot unmarshal string into Go value of type common.OpenAPIDefinition"),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_crds() string {
	return `
data "kustomization_overlay" "test" {
	crds = [
		"test_kustomizations/crd/initial/crd.yaml",
	]
}
`
}

//
//
// Test images attr
func TestDataSourceKustomizationOverlay_images(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationImagesConfig(),
				Check:  resource.TestCheckOutput("check", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-basic\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"image\":\"testname@sha256:abcdefghijklmnop123456\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}"),
			},
		},
	})
}

func testKustomizationImagesConfig() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"test_kustomizations/basic/initial",
	]

	images {
		name = "nginx"
		new_name = "testname"
		new_tag = "testtag"
		digest = "sha256:abcdefghijklmnop123456"
	}
}

output "check" {
	value = data.kustomization_overlay.test.manifests["apps_v1_Deployment|test-basic|test"]
}
`
}

//
//
// Test replicas attr
func TestDataSourceKustomizationOverlay_replicas(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationReplicasConfig(),
				Check:  resource.TestCheckOutput("check", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-basic\"},\"spec\":{\"replicas\":5,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}"),
			},
		},
	})
}

func testKustomizationReplicasConfig() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"test_kustomizations/basic/initial",
	]

	replicas {
		name = "test"
		count = 5
	}
}

output "check" {
	value = data.kustomization_overlay.test.manifests["apps_v1_Deployment|test-basic|test"]
}
`
}
