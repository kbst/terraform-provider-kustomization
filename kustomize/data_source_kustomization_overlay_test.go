package kustomize

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

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
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "labels.%", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "components.#", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "config_map_generator.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "crds.#", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "generators.#", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "generator_options.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "images.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "namespace", "test-overlay-basic"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "replicas.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "resources.#", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "secret_generator.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "patches.#", "1"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "transformers.#", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "vars.#", "1"),

					// Generated
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids.#", "0"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids_prio.#", "3"),
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

	labels {
		pairs = {}
	}

	components = []

	config_map_generator {}

	crds = []

	generators = []

	generator_options {}

	images {}

	name_prefix = "test-"

	namespace = "test-overlay-basic"

	name_suffix = "-test"

	replicas {}

	resources = []

	secret_generator {}

	patches {}

	transformers = []

	vars {}
}
`
}

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
	value = data.kustomization_overlay.test.manifests["_/Namespace/_/test-basic"]
}
`
}

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
	value = data.kustomization_overlay.test.manifests["_/Namespace/_/test-basic"]
}
`
}

// Test labels attr
func TestDataSourceKustomizationOverlay_labels(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_labels(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\",\"test-label\":\"true\",\"test-label-selector\":\"true\"},\"name\":\"test\",\"namespace\":\"test-basic\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\",\"test-label-selector\":\"true\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\",\"test-label-selector\":\"true\"}},\"spec\":{\"containers\":[{\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_labels() string {
	return `
data "kustomization_overlay" "test" {
	labels {
		pairs = {
			test-label = true
		}
	}
	labels {
		pairs = {
			test-label-selector = true
		}
		include_selectors = true
	}

	resources = [
		"test_kustomizations/basic/initial",
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/test-basic/test"]
}
`
}

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
	value = data.kustomization_overlay.test.manifests["_/ConfigMap/_/from-component-6ct58987ht"]
}
`
}

// Test config_map_generator attr
func TestDataSourceKustomizationOverlay_configMapGenerator(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationConfigMapGeneratorConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_cm1", "{\"apiVersion\":\"v1\",\"data\":{\"KEY1\":\"VALUE1\",\"KEY2\":\"VALUE2\"},\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"test-configmap1\"}}"),
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

		options {
			disable_name_suffix_hash = true
		}
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
	value = data.kustomization_overlay.test.manifests["_/ConfigMap/_/test-configmap1"]
}

output "check_cm2" {
	value = data.kustomization_overlay.test.manifests["_/ConfigMap/_/test-configmap2-5tgmgc9cmf"]
}
`
}

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
	value = data.kustomization_overlay.test.manifests["_/Namespace/_/test-overlay-namespace"]
}

`
}

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
	value = data.kustomization_overlay.test.manifests["_/Service/test-basic/test-test-test"]
}

`
}

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
	value = data.kustomization_overlay.test.manifests["_/Namespace/_/test-basic"]
}

`
}

// Test transformers attr
func TestDataSourceKustomizationOverlay_transformers(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_transformers(),
				Check:  resource.TestCheckOutput("check", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\",\"test.example.com/test-label\":\"test-value\"},\"name\":\"test\",\"namespace\":\"test-transformer-config\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}"),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_transformers() string {
	return `
data "kustomization_overlay" "test" {
	transformers = [
		"test_kustomizations/transformer_configs/modified/label.yaml",
	]

	resources = [
		"test_kustomizations/transformer_configs/initial",
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/test-transformer-config/test"]
}
`
}

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
				ExpectError: regexp.MustCompile("json: cannot unmarshal string into Go value of type accumulator.OpenAPIDefinition"),
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

// Test generators attr
func TestDataSourceKustomizationOverlay_generators(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_generators(),
				Check:  resource.TestCheckOutput("check", "{\"apiVersion\":\"v1\",\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"testcm-6ct58987ht\"}}"),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_generators() string {
	return `
data "kustomization_overlay" "test" {
	generators = [
		"test_kustomizations/_test_files/cmGenerator.yaml"
	]
}

output "check" {
	value = data.kustomization_overlay.test.manifests["_/ConfigMap/_/testcm-6ct58987ht"]
}
`
}

// Test generator_options attr
func TestDataSourceKustomizationOverlay_generator_options(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationConfig_generator_options(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_configmap", "{\"apiVersion\":\"v1\",\"kind\":\"ConfigMap\",\"metadata\":{\"annotations\":{\"test-annotation\":\"test\"},\"labels\":{\"test-label\":\"test\"},\"name\":\"test-configmap\"}}"),
					resource.TestCheckOutput("check_secret", "{\"apiVersion\":\"v1\",\"data\":{},\"kind\":\"Secret\",\"metadata\":{\"annotations\":{\"test-annotation\":\"test\"},\"labels\":{\"test-label\":\"test\"},\"name\":\"test-secret\"},\"type\":\"Opaque\"}"),
				),
			},
		},
	})
}

func testKustomizationConfig_generator_options() string {
	return `
data "kustomization_overlay" "test" {
	generator_options {
		labels = {
			test-label = "test"
		}

		annotations = {
			test-annotation = "test"
		}

		disable_name_suffix_hash = true
	}

	config_map_generator {
		name = "test-configmap"
		literals = []
	}

	secret_generator {
		name = "test-secret"
		literals = []
	}
}

output "check_configmap" {
	value = data.kustomization_overlay.test.manifests["_/ConfigMap/_/test-configmap"]
}

output "check_secret" {
	value = data.kustomization_overlay.test.manifests["_/Secret/_/test-secret"]
}
`
}

// Test images attr
func TestDataSourceKustomizationOverlay_images(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationImagesConfig(),
				Check:  resource.TestCheckOutput("check", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-basic\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"image\":\"testname:testtag@sha256:abcdefghijklmnop123456\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}"),
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
	value = data.kustomization_overlay.test.manifests["apps/Deployment/test-basic/test"]
}
`
}

// Test patches attr
func TestDataSourceKustomizationOverlay_patches(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationPatchesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_dep", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-basic\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"TESTENV\",\"value\":\"true\"}],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}"),
					resource.TestCheckOutput("check_ingress", "{\"apiVersion\":\"networking.k8s.io/v1\",\"kind\":\"Ingress\",\"metadata\":{\"annotations\":{\"nginx.ingress.kubernetes.io/rewrite-target\":\"/\"},\"name\":\"test\",\"namespace\":\"test-basic\"},\"spec\":{\"rules\":[{\"http\":{\"paths\":[{\"backend\":{\"service\":{\"name\":\"test\",\"port\":{\"number\":80}}},\"path\":\"/testpath\",\"pathType\":\"Prefix\"}]}}]}}"),
				),
			},
		},
	})
}

func testKustomizationPatchesConfig() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"test_kustomizations/basic/initial",
	]

	patches {
		path = "test_kustomizations/_test_files/deployment_patch_env.yaml"
		target {
			label_selector = "app=test"
		}
	}

	patches {
		patch = <<-EOF
			- op: replace
			  path: /spec/rules/0/http/paths/0/path
			  value: /newpath
		EOF
		target {
			group = "networking.k8s.io"
			version = "v1beta1"
			kind = "Ingress"
			name = "test"
			namespace = "test-basic"
			annotation_selector = "nginx.ingress.kubernetes.io/rewrite-target"
		}
	}
}

output "check_dep" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/test-basic/test"]
}

output "check_ingress" {
	value = data.kustomization_overlay.test.manifests["networking.k8s.io/Ingress/test-basic/test"]
}
`
}

// Test replacements attr
func TestDataSourceKustomizationOverlay_replacements(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationReplacementsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_target1", "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"replacement-pod\",\"namespace\":\"test-replacements\"},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"MODIFY_ME\",\"value\":\"this-is-replace1\"},{\"name\":\"LEAVE_ME_ALONE\",\"value\":\"this should stay untouched\"}],\"image\":\"postgres:latest\",\"name\":\"modify-me\"},{\"env\":[{\"name\":\"UNMODIFIED\",\"value\":\"still the same\"}],\"image\":\"nginx:latest\",\"name\":\"leave-me-alone\"}]}}"),
					resource.TestCheckOutput("check_target2", "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"replacement-pod-2\",\"namespace\":\"test-replacements\"},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"MODIFY_ME\",\"value\":\"this should be ignored by reject rule\"},{\"name\":\"LEAVE_ME_ALONE\",\"value\":\"this should stay untouched\"}],\"image\":\"postgres:latest\",\"name\":\"modify-me\"},{\"env\":[{\"name\":\"UNMODIFIED\",\"value\":\"still the same\"}],\"image\":\"nginx:latest\",\"name\":\"leave-me-alone\"}]}}"),
					resource.TestCheckOutput("check_target3", "{\"apiVersion\":\"v1\",\"data\":{\"REPLACE_ME\":\"this-is-replace1\"},\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"replacement-target\",\"namespace\":\"test-replacements\"}}"),
				),
			},
		},
	})
}

func testKustomizationReplacementsConfig() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"test_kustomizations/replacements",
	]

	replacements {
		source {
			kind = "ConfigMap"
			name = "replacement-source"
			field_path = "data.replace1"
		}
		target {
			select {
				kind = "Pod"
			}
			reject {
				name = "replacement-pod-2"
			}
			field_paths = ["spec.containers.[name=modify-me].env.[name=MODIFY_ME].value"]
		}
		target {
			select {
				kind = "ConfigMap"
				name = "replacement-target"
			}
			field_paths = ["data.REPLACE_ME"]
		}
	}
}

output "check_target1" {
	value = data.kustomization_overlay.test.manifests["_/Pod/test-replacements/replacement-pod"]
}

output "check_target2" {
	value = data.kustomization_overlay.test.manifests["_/Pod/test-replacements/replacement-pod-2"]
}

output "check_target3" {
	value = data.kustomization_overlay.test.manifests["_/ConfigMap/test-replacements/replacement-target"]
}
`
}

func TestDataSourceKustomizationOverlay_replacements_from_file(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationReplacementsFromFileConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_target1", "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"replacement-pod\",\"namespace\":\"test-replacements\"},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"MODIFY_ME\",\"value\":\"this-is-replace1\"},{\"name\":\"LEAVE_ME_ALONE\",\"value\":\"this should stay untouched\"}],\"image\":\"postgres:latest\",\"name\":\"modify-me\"},{\"env\":[{\"name\":\"UNMODIFIED\",\"value\":\"still the same\"}],\"image\":\"nginx:latest\",\"name\":\"leave-me-alone\"}]}}"),
					resource.TestCheckOutput("check_target2", "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"replacement-pod-2\",\"namespace\":\"test-replacements\"},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"MODIFY_ME\",\"value\":\"this should be ignored by reject rule\"},{\"name\":\"LEAVE_ME_ALONE\",\"value\":\"this should stay untouched\"}],\"image\":\"postgres:latest\",\"name\":\"modify-me\"},{\"env\":[{\"name\":\"UNMODIFIED\",\"value\":\"still the same\"}],\"image\":\"nginx:latest\",\"name\":\"leave-me-alone\"}]}}"),
					resource.TestCheckOutput("check_target3", "{\"apiVersion\":\"v1\",\"data\":{\"REPLACE_ME\":\"this-is-replace1\"},\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"replacement-target\",\"namespace\":\"test-replacements\"}}"),
				),
			},
		},
	})
}

func testKustomizationReplacementsFromFileConfig() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"test_kustomizations/replacements",
	]

	replacements {
		path = "test_kustomizations/replacements/replacements.yaml"
	}
}

output "check_target1" {
	value = data.kustomization_overlay.test.manifests["_/Pod/test-replacements/replacement-pod"]
}

output "check_target2" {
	value = data.kustomization_overlay.test.manifests["_/Pod/test-replacements/replacement-pod-2"]
}

output "check_target3" {
	value = data.kustomization_overlay.test.manifests["_/ConfigMap/test-replacements/replacement-target"]
}
`
}

func TestDataSourceKustomizationOverlay_replacements_configmap(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationReplacementsConfigMapConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_target1", "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"replacement-pod\",\"namespace\":\"test-replacements\"},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"MODIFY_ME\",\"value\":\"a generated value\"},{\"name\":\"LEAVE_ME_ALONE\",\"value\":\"this should stay untouched\"}],\"image\":\"postgres:latest\",\"name\":\"modify-me\"},{\"env\":[{\"name\":\"UNMODIFIED\",\"value\":\"still the same\"}],\"image\":\"nginx:latest\",\"name\":\"leave-me-alone\"}]}}"),
					resource.TestCheckOutput("check_source", "{\"apiVersion\":\"v1\",\"data\":{\"replace1\":\"a generated value\"},\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"replacement-generated\",\"namespace\":\"test-replacements\"}}"),
				),
			},
		},
	})
}

func testKustomizationReplacementsConfigMapConfig() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"test_kustomizations/replacements",
	]

	config_map_generator {
		name = "replacement-generated"
		namespace = "test-replacements"
		literals = [
			"replace1=a generated value"
		]
		options {
			disable_name_suffix_hash = true
		}
	}

	replacements {
		source {
			kind = "ConfigMap"
			name = "replacement-generated"
			field_path = "data.replace1"
		}
		target {
			select {
				kind = "Pod"
			}
			reject {
				name = "replacement-pod-2"
			}
			field_paths = ["spec.containers.[name=modify-me].env.[name=MODIFY_ME].value"]
		}
	}
}

output "check_target1" {
	value = data.kustomization_overlay.test.manifests["_/Pod/test-replacements/replacement-pod"]
}

output "check_source" {
	value = data.kustomization_overlay.test.manifests["_/ConfigMap/test-replacements/replacement-generated"]
}
`
}

func TestDataSourceKustomizationOverlay_replacements_patch(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationReplacementsPatchConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_target1", "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"replacement-pod\",\"namespace\":\"test-replacements\"},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"MODIFY_ME\",\"value\":\"a patched value\"},{\"name\":\"LEAVE_ME_ALONE\",\"value\":\"this should stay untouched\"}],\"image\":\"postgres:latest\",\"name\":\"modify-me\"},{\"env\":[{\"name\":\"UNMODIFIED\",\"value\":\"still the same\"}],\"image\":\"nginx:latest\",\"name\":\"leave-me-alone\"}]}}"),
					resource.TestCheckOutput("check_source", "{\"apiVersion\":\"v1\",\"data\":{\"REPLACE_ME\":\"this should stay untouched\",\"replace1\":\"a patched value\",\"replace2\":\"this-is-replace2\"},\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"replacement-source\",\"namespace\":\"test-replacements\"}}"),
				),
			},
		},
	})
}

func testKustomizationReplacementsPatchConfig() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"test_kustomizations/replacements",
	]

	patches {
		target {
			name = "replacement-source"
			namespace = "test-replacements"
		}
		patch = <<-EOF
		  - op: replace
		    path: /data/replace1
		    value: a patched value
		EOF
	}

	replacements {
		source {
			kind = "ConfigMap"
			name = "replacement-source"
			field_path = "data.replace1"
		}
		target {
			select {
				kind = "Pod"
			}
			reject {
				name = "replacement-pod-2"
			}
			field_paths = ["spec.containers.[name=modify-me].env.[name=MODIFY_ME].value"]
		}
	}
}

output "check_target1" {
	value = data.kustomization_overlay.test.manifests["_/Pod/test-replacements/replacement-pod"]
}

output "check_source" {
	value = data.kustomization_overlay.test.manifests["_/ConfigMap/test-replacements/replacement-source"]
}
`
}

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
	value = data.kustomization_overlay.test.manifests["apps/Deployment/test-basic/test"]
}
`
}

// Test secret_generator attr
func TestDataSourceKustomizationOverlay_secretGenerator(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationSecretGeneratorConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_cm1", "{\"apiVersion\":\"v1\",\"data\":{\"KEY1\":\"VkFMVUUx\",\"KEY2\":\"VkFMVUUy\"},\"kind\":\"Secret\",\"metadata\":{\"name\":\"test-secret1\"},\"type\":\"Opaque\"}"),
					resource.TestCheckOutput("check_cm2", "{\"apiVersion\":\"v1\",\"data\":{\"ENV1\":\"VkFMVUUx\",\"ENV2\":\"VkFMVUUy\",\"KEY1\":\"VkFMVUUx\",\"KEY2\":\"VkFMVUUy\",\"properties.env\":\"RU5WMT1WQUxVRTEKRU5WMj1WQUxVRTIK\"},\"kind\":\"Secret\",\"metadata\":{\"name\":\"test-secret2-h55cfd6gfg\"},\"type\":\"Opaque\"}"),
				),
			},
		},
	})
}

func testKustomizationSecretGeneratorConfig() string {
	return `
data "kustomization_overlay" "test" {
	secret_generator {
		name = "test-secret1"
		type = "Opaque"
		literals = [
			"KEY1=VALUE1",
			"KEY2=VALUE2"
		]

		options {
			disable_name_suffix_hash = true
		}
	}

	secret_generator {
		name = "test-secret2"
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
	value = data.kustomization_overlay.test.manifests["_/Secret/_/test-secret1"]
}

output "check_cm2" {
	value = data.kustomization_overlay.test.manifests["_/Secret/_/test-secret2-h55cfd6gfg"]
}
`
}

// Test vars attr
func TestDataSourceKustomizationOverlay_vars(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationVarsConfig(),
				Check:  resource.TestCheckOutput("check", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-basic\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"TESTENV\",\"value\":\"test-basic\"}],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}"),
			},
		},
	})
}

func testKustomizationVarsConfig() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"test_kustomizations/basic/initial",
	]

	vars {
		name = "TEST_VAR_NAMESPACE"
		obj_ref {
			api_version = "v1"
			kind = "Service"
			name = "test"
		}
		field_ref {
			field_path = "metadata.namespace"
		}
	}

	patches {
		patch = <<-EOF
			- op: add
			  path: /spec/template/spec/containers/0/env
			  value: [{"name": "TESTENV", "value": "$(TEST_VAR_NAMESPACE)"}]
		EOF
		target {
			group = "apps"
			version = "v1"
			kind = "Deployment"
			name = "test"
		}
	}
}

output "check" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/test-basic/test"]
}
`
}

func TestDataSourceKustomizationOverlay_conflict(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	for _, n := range konfig.RecognizedKustomizationFileNames() {
		fSys.WriteFile(n, []byte{})

		err := refuseExistingKustomization(fSys)
		assert.EqualErrorf(
			t,
			err,
			fmt.Sprintf("buildKustomizeOverlay: Can not build dynamic overlay, found %q in working directory.", n),
			"",
			nil,
		)

		fSys.RemoveAll(n)
	}
}

// Test patch options attr
func TestDataSourceKustomizationOverlay_patchOptionsNoop(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationPatchOptionsConfigNoop(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_ns", `{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"test-basic"}}`),
				),
			},
		},
	})
}

func testKustomizationPatchOptionsConfigNoop() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"test_kustomizations/basic/initial",
	]
	patches {
		target {
			kind = "Namespace"
			name = "test-basic"
		}
		patch = <<-EOF
			kind: Namespace
			metadata:
			  name: new-basic
		EOF
		options {
			allow_name_change = false
		}
	}
}

output "check_ns" {
	value = data.kustomization_overlay.test.manifests["_/Namespace/_/test-basic"]
}
`
}

// Test patch options attr(
func TestDataSourceKustomizationOverlay_patchOptions(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationPatchOptionsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_ns", `{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"new-basic"}}`),
				),
			},
		},
	})
}

func testKustomizationPatchOptionsConfig() string {
	return `
data "kustomization_overlay" "test" {
	resources = [
		"test_kustomizations/basic/initial",
	]
	patches {
		target {
			kind = "Namespace"
			name = "test-basic"
		}
		patch = <<-EOF
			kind: Namespace
			metadata:
			  name: new-basic
		EOF
		options {
			allow_name_change = true
		}
	}
}

output "check_ns" {
	value = data.kustomization_overlay.test.manifests["_/Namespace/_/new-basic"]
}
`
}

// Test helm with common_annotations attr
func TestDataSourceKustomizationOverlay_commonAnnotations_helm(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_commonAnnotations_helm(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"annotations\":{\"test-annotation\":\"true\"},\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\",\"namespace\":\"test-basic\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":80}],\"selector\":{\"app\":\"nginx\"},\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_commonAnnotations_helm() string {
	return `
data "kustomization_overlay" "test" {
	common_annotations = {
		test-annotation: true
	}

	resources = [
		"test_kustomizations/helm/initial",
	]

	kustomize_options {
		enable_helm = true
		helm_path = "helm"
	}
}

output "check" {
	value = data.kustomization_overlay.test.manifests["_/Service/test-basic/nginx"]
}
`
}

// Test helm_charts with kustomized namespace inclusion
func TestDataSourceKustomizationOverlay_helm_charts_attr(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_helm_charts_attr(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("service", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\",\"namespace\":\"test-basic\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":80}],\"selector\":{\"app\":\"nginx\"},\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}"),
					resource.TestCheckOutput("deployment", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\",\"namespace\":\"test-basic\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"nginx\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"}},\"spec\":{\"containers\":[{\"image\":\"nginx:6.0.10\",\"name\":\"test-basic\",\"resources\":{}}]}}},\"status\":{}}"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids.#", "3"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids_prio.#", "3"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "manifests.%", "3"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_helm_charts_attr() string {
	return `
data "kustomization_overlay" "test" {
	helm_globals {
		chart_home = "./test_kustomizations/helm/initial/charts/"
	}

	helm_charts {
		name = "test-basic"
		version = "0.0.1"
		namespace = "not-used"
	}

	namespace = "test-basic"

	kustomize_options {
		enable_helm = true
		helm_path = "helm"
	}
}

output "service" {
	value = data.kustomization_overlay.test.manifests["_/Service/test-basic/nginx"]
}

output "deployment" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/test-basic/nginx"]
}

`
}

// Test helm_charts release_name
// if release_name were not set, there would be many RELEASE-NAME values in the rendered template
func TestDataSourceKustomizationOverlay_helm_charts_releasename(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_helm_charts_releasename(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("service", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"my-release\"},\"name\":\"my-release\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":80}],\"selector\":{\"app\":\"my-release\"},\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}"),
					resource.TestCheckOutput("deployment", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"my-release\"},\"name\":\"my-release\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"my-release\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"my-release\"}},\"spec\":{\"containers\":[{\"image\":\"nginx:6.0.10\",\"name\":\"test-releasename\",\"resources\":{}}]}}},\"status\":{}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_helm_charts_releasename() string {
	return `
data "kustomization_overlay" "test" {
	helm_globals {
		chart_home = "./test_kustomizations/helm/initial/charts/"
	}

	helm_charts {
		name = "test-releasename"
		version = "0.0.1"
		release_name = "my-release"
	}

	kustomize_options {
		enable_helm = true
		helm_path = "helm"
	}
}

output "service" {
	value = data.kustomization_overlay.test.manifests["_/Service/_/my-release"]
}

output "deployment" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/_/my-release"]
}
`
}

// Test helm_charts values_file
// values_file overrides the default values that accompany the chart
func TestDataSourceKustomizationOverlay_helm_charts_valuesFile(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_helm_charts_valuesFile(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// default values:
					//   replicas = 1, port = 80,  image name = nginx
					// overridden in alt-values.yaml
					//   replicas = 2, port = 443, image name = my-nginx
					resource.TestCheckOutput("service", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":443,\"protocol\":\"TCP\",\"targetPort\":443}],\"selector\":{\"app\":\"nginx\"},\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}"),
					resource.TestCheckOutput("deployment", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\"},\"spec\":{\"replicas\":2,\"selector\":{\"matchLabels\":{\"app\":\"nginx\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"}},\"spec\":{\"containers\":[{\"image\":\"my-nginx:6.0.10\",\"name\":\"test-basic\",\"resources\":{}}]}}},\"status\":{}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_helm_charts_valuesFile() string {
	return `
data "kustomization_overlay" "test" {
	helm_globals {
		chart_home = "./test_kustomizations/helm/initial/charts/"
	}

	helm_charts {
		name = "test-basic"
		version = "0.0.1"
		values_file = "./test_kustomizations/helm/initial/alt-values.yaml"
	}

	kustomize_options {
		enable_helm = true
		helm_path = "helm"
	}
}

output "service" {
	value = data.kustomization_overlay.test.manifests["_/Service/_/nginx"]
}

output "deployment" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/_/nginx"]
}
`
}

// Test helm_charts values_inline
// values_inline overrides the default values that accompany the chart, directly from HCL
func TestDataSourceKustomizationOverlay_helm_charts_valuesInline(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_helm_charts_valuesInline(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// default values:
					//   replicas = 1, port = 80,  image name = nginx
					// overridden to:
					//   replicas = 2, port = 443, image name = my-nginx
					resource.TestCheckOutput("service", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":443,\"protocol\":\"TCP\",\"targetPort\":443}],\"selector\":{\"app\":\"nginx\"},\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}"),
					resource.TestCheckOutput("deployment", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\"},\"spec\":{\"replicas\":2,\"selector\":{\"matchLabels\":{\"app\":\"nginx\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"}},\"spec\":{\"containers\":[{\"image\":\"my-nginx:6.0.10\",\"name\":\"test-basic\",\"resources\":{}}]}}},\"status\":{}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_helm_charts_valuesInline() string {
	return `
data "kustomization_overlay" "test" {
	helm_globals {
		chart_home = "./test_kustomizations/helm/initial/charts/"
	}

	helm_charts {
		name = "test-basic"
		version = "0.0.1"
		values_inline = <<VALUES
      replicaCount: 2

      image:
        repository: my-nginx

      nginx:
        port: 443
    VALUES
	}

	kustomize_options {
		enable_helm = true
		helm_path = "helm"
	}
}

output "service" {
	value = data.kustomization_overlay.test.manifests["_/Service/_/nginx"]
}

output "deployment" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/_/nginx"]
}
`
}

// Test helm_charts values_merge
// values_merge determines how to merge values if both values_files and values_inline are used simultaneously
func TestDataSourceKustomizationOverlay_helm_charts_valuesMerge(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_helm_charts_valuesMerge(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// values_file specifies:
					//   image name = my-nginx
					//   image tag = "6.0.10"
					// values_inline specifies:
					//   replicas = 3
					//   image tag = "7.0.0"
					//   port =-443
					// merged to:
					//   replicas = 3
					//   port = 443
					//   image name = my-nginx
					//   image tag = "7.0.0"
					resource.TestCheckOutput("service", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":443,\"protocol\":\"TCP\",\"targetPort\":443}],\"selector\":{\"app\":\"nginx\"},\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}"),
					resource.TestCheckOutput("deployment", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\"},\"spec\":{\"replicas\":3,\"selector\":{\"matchLabels\":{\"app\":\"nginx\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"}},\"spec\":{\"containers\":[{\"image\":\"my-nginx:7.0.0\",\"name\":\"test-basic\",\"resources\":{}}]}}},\"status\":{}}"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_helm_charts_valuesMerge() string {
	return `
data "kustomization_overlay" "test" {
	helm_globals {
		chart_home = "./test_kustomizations/helm/initial/charts/"
	}

	helm_charts {
		name = "test-basic"
		version = "0.0.1"
		values_file = "./test_kustomizations/helm/initial/merge-values.yaml"
		values_inline = <<VALUES
      replicaCount: 3
      image:
        tag: 7.0.0
      nginx:
        port: 443
    VALUES
		values_merge = "override"
	}

	kustomize_options {
		enable_helm = true
		helm_path = "helm"
	}
}

output "service" {
	value = data.kustomization_overlay.test.manifests["_/Service/_/nginx"]
}

output "deployment" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/_/nginx"]
}
`
}

// Test helm_charts include_crds
// this uses the same chart as TestDataSourceKustomizationOverlay_helm_charts_attr
// only enabling the include_crds attribute as well
func TestDataSourceKustomizationOverlay_helm_charts_includeCrds(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_helm_charts_includeCrds(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("service", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":80}],\"selector\":{\"app\":\"nginx\"},\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}"),
					resource.TestCheckOutput("deployment", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"nginx\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"}},\"spec\":{\"containers\":[{\"image\":\"nginx:6.0.10\",\"name\":\"test-basic\",\"resources\":{}}]}}},\"status\":{}}"),
					resource.TestCheckOutput("crd", "{\"apiVersion\":\"apiextensions.k8s.io/v1\",\"kind\":\"CustomResourceDefinition\",\"metadata\":{\"name\":\"crontabs.stable.example.com\"},\"spec\":{\"group\":\"stable.example.com\",\"names\":{\"kind\":\"CronTab\",\"plural\":\"crontabs\",\"shortNames\":[\"ct\"],\"singular\":\"crontab\"},\"scope\":\"Namespaced\",\"versions\":[{\"name\":\"v1\",\"schema\":{\"openAPIV3Schema\":{\"properties\":{\"spec\":{\"properties\":{\"cronSpec\":{\"type\":\"string\"},\"image\":{\"type\":\"string\"},\"replicas\":{\"type\":\"integer\"}},\"type\":\"object\"}},\"type\":\"object\"}},\"served\":true,\"storage\":true}]}}"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids.#", "4"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids_prio.#", "3"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "manifests.%", "4"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_helm_charts_includeCrds() string {
	return `
data "kustomization_overlay" "test" {
	helm_globals {
		chart_home = "./test_kustomizations/helm/initial/charts/"
	}

	helm_charts {
		name = "test-basic"
		version = "0.0.1"
		include_crds = true
	}

	kustomize_options {
		enable_helm = true
		helm_path = "helm"
	}
}

output "service" {
	value = data.kustomization_overlay.test.manifests["_/Service/_/nginx"]
}

output "deployment" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/_/nginx"]
}

output "crd" {
	value = data.kustomization_overlay.test.manifests["apiextensions.k8s.io/CustomResourceDefinition/_/crontabs.stable.example.com"]
}
`
}

// Test helm_charts multiple charts in one kustomization overlay
func TestDataSourceKustomizationOverlay_helm_charts_multiple_charts(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_helm_charts_multiple_charts(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("service_1", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":80}],\"selector\":{\"app\":\"nginx\"},\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}"),
					resource.TestCheckOutput("deployment_1", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"nginx\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"nginx\"}},\"spec\":{\"containers\":[{\"image\":\"nginx:6.0.10\",\"name\":\"test-basic\",\"resources\":{}}]}}},\"status\":{}}"),
					resource.TestCheckOutput("service_2", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"my-release\"},\"name\":\"my-release\"},\"spec\":{\"ports\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":80}],\"selector\":{\"app\":\"my-release\"},\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}"),
					resource.TestCheckOutput("deployment_2", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"my-release\"},\"name\":\"my-release\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"my-release\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"my-release\"}},\"spec\":{\"containers\":[{\"image\":\"nginx:6.0.10\",\"name\":\"test-releasename\",\"resources\":{}}]}}},\"status\":{}}"),
					// 3 manifests from first chart, 3 from second chart
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids.#", "6"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "manifests.%", "6"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_helm_charts_multiple_charts() string {
	return `
data "kustomization_overlay" "test" {
	helm_globals {
		chart_home = "./test_kustomizations/helm/initial/charts/"
	}

	helm_charts {
		name = "test-releasename"
		version = "0.0.1"
		release_name = "my-release"
	}

	helm_charts {
		name = "test-basic"
		version = "0.0.1"
	}

	kustomize_options {
		enable_helm = true
		helm_path = "helm"
	}
}

output "service_1" {
	value = data.kustomization_overlay.test.manifests["_/Service/_/nginx"]
}

output "deployment_1" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/_/nginx"]
}

output "service_2" {
	value = data.kustomization_overlay.test.manifests["_/Service/_/my-release"]
}

output "deployment_2" {
	value = data.kustomization_overlay.test.manifests["apps/Deployment/_/my-release"]
}
`
}

// Test helm_charts remote repo
func TestDataSourceKustomizationOverlay_helm_charts_repo(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKustomizationOverlayConfig_helm_charts_repo(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("rolebinding", "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"RoleBinding\",\"metadata\":{\"labels\":{\"app\":\"kube-prometheus-stack-alertmanager\",\"app.kubernetes.io/instance\":\"my-release\",\"app.kubernetes.io/managed-by\":\"Helm\",\"app.kubernetes.io/part-of\":\"kube-prometheus-stack\",\"app.kubernetes.io/version\":\"23.3.2\",\"chart\":\"kube-prometheus-stack-23.3.2\",\"heritage\":\"Helm\",\"release\":\"my-release\"},\"name\":\"my-release-kube-prometheus-alertmanager\",\"namespace\":\"default\"},\"roleRef\":{\"apiGroup\":\"rbac.authorization.k8s.io\",\"kind\":\"Role\",\"name\":\"my-release-kube-prometheus-alertmanager\"},\"subjects\":[{\"kind\":\"ServiceAccount\",\"name\":\"my-release-kube-prometheus-alertmanager\",\"namespace\":\"default\"}]}"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "ids.#", "136"),
					resource.TestCheckResourceAttr("data.kustomization_overlay.test", "manifests.%", "136"),
				),
			},
		},
	})
}

func testDataSourceKustomizationOverlayConfig_helm_charts_repo() string {
	return `
data "kustomization_overlay" "test" {
	helm_globals {
		chart_home = "./test_kustomizations/helm/remote/"
	}
	helm_charts {
		name = "kube-prometheus-stack"
		repo = "https://prometheus-community.github.io/helm-charts"
		version = "23.3.2"
		release_name = "my-release"
	}

	kustomize_options {
		enable_helm = true
		helm_path = "helm"
	}
}

output "rolebinding" {
	value = data.kustomization_overlay.test.manifests["rbac.authorization.k8s.io/RoleBinding/default/my-release-kube-prometheus-alertmanager"]
}
`
}
