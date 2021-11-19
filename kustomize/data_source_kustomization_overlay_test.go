package kustomize

import (
	"fmt"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kyaml/filesys"
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
	value = data.kustomization_overlay.test.manifests["_/Namespace/_/test-basic"]
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
	value = data.kustomization_overlay.test.manifests["_/Namespace/_/test-basic"]
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
	value = data.kustomization_overlay.test.manifests["_/ConfigMap/_/from-component-6ct58987ht"]
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
	value = data.kustomization_overlay.test.manifests["_/Namespace/_/test-overlay-namespace"]
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
	value = data.kustomization_overlay.test.manifests["_/Service/test-basic/test-test-test"]
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
	value = data.kustomization_overlay.test.manifests["_/Namespace/_/test-basic"]
}

`
}

//
//
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

//
//
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

//
//
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
	value = data.kustomization_overlay.test.manifests["apps/Deployment/test-basic/test"]
}
`
}

//
//
// Test patches attr
func TestDataSourceKustomizationOverlay_patches(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationPatchesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_dep", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-basic\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"TESTENV\",\"value\":\"true\"}],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}"),
					resource.TestCheckOutput("check_ingress", "{\"apiVersion\":\"networking.k8s.io/v1beta1\",\"kind\":\"Ingress\",\"metadata\":{\"annotations\":{\"nginx.ingress.kubernetes.io/rewrite-target\":\"/\"},\"name\":\"test\",\"namespace\":\"test-basic\"},\"spec\":{\"rules\":[{\"http\":{\"paths\":[{\"backend\":{\"serviceName\":\"test\",\"servicePort\":80},\"path\":\"/newpath\"}]}}]}}"),
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
		target = {
			label_selector = "app=test"
		}
	}

	patches {
		patch = <<-EOF
			- op: replace
			  path: /spec/rules/0/http/paths/0/path
			  value: /newpath
		EOF
		target = {
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
	value = data.kustomization_overlay.test.manifests["apps/Deployment/test-basic/test"]
}
`
}

//
//
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

//
//
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
		obj_ref = {
			api_version = "v1"
			kind = "Service"
			name = "test"
		}
		field_ref = {
			field_path = "metadata.namespace"
		}
	}

	patches {
		patch = <<-EOF
			- op: add
			  path: /spec/template/spec/containers/0/env
			  value: [{"name": "TESTENV", "value": "$(TEST_VAR_NAMESPACE)"}]
		EOF
		target = {
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

//
//
// Test module
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

//
//
// Test module
func TestDataSourceKustomizationOverlay_module(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationOverlayConfig_module(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check_dep", "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{\"test-annotation\":\"true\"},\"labels\":{\"app\":\"test\",\"test-label\":\"true\"},\"name\":\"tp-test-ts\",\"namespace\":\"test-module\"},\"spec\":{\"replicas\":9,\"selector\":{\"matchLabels\":{\"app\":\"test\",\"test-label\":\"true\"}},\"strategy\":{},\"template\":{\"metadata\":{\"annotations\":{\"test-annotation\":\"true\"},\"labels\":{\"app\":\"test\",\"test-label\":\"true\"}},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"TESTENV\",\"value\":\"true\"},{\"name\":\"TESTVAR\",\"value\":\"tp-os-ts-46f8b28mk5\"}],\"image\":\"test\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}"),
					resource.TestCheckOutput("check_cm", "{\"apiVersion\":\"v1\",\"data\":{\"KEY1\":\"VALUE1\"},\"kind\":\"ConfigMap\",\"metadata\":{\"annotations\":{\"kustomize_generated\":\"true\",\"test-annotation\":\"true\"},\"labels\":{\"test-label\":\"true\"},\"name\":\"tp-ocm-ts\",\"namespace\":\"test-module\"}}"),
				),
			},
		},
	})
}

func testKustomizationOverlayConfig_module() string {
	absPath, _ := filepath.Abs("test_module")
	modulePath := filepath.ToSlash(absPath)
	return fmt.Sprintf(`
module "test" {
	source = %q

	common_annotations = {
		test-annotation = "true"
	}

	common_labels = {
		test-label = "true"
	}

	config_map_generator = [{
		name = "ocm"
		literals = [
			"KEY1=VALUE1"
		]
		options = {
			disable_name_suffix_hash = true
		}
	}]

	generator_options = {
		annotations = {
			"kustomize_generated" = "true"
		}
	}

	images = [{
		name = "nginx"
		new_name = "test"
	}]

	name_prefix = "tp-"
	namespace = "test-module"
	name_suffix = "-ts"

	patches = [
		{
			path = "%s/../test_kustomizations/_test_files/deployment_patch_env.yaml"
			patch = null
			target = {}
		}, {
			path = null
			patch = <<-EOF
				- op: add
				  path: /spec/template/spec/containers/0/env/-
				  value: {"name": "TESTVAR", "value": "$(TEST_VAR)"}
			EOF
			target = {
				group = "apps"
				version = "v1"
				kind = "Deployment"
				name = "test"
			}
		}
	]

	replicas = [{
		name = "test"
		count = 9
	}]

	secret_generator = [{
		name = "os"
	}]

	vars = [{
		name = "TEST_VAR"
		obj_ref = {
			api_version = "v1"
			kind = "Secret"
			name = "os"
		}
	}]
}

output "check_dep" {
	value = module.test.kustomization.manifests["apps/Deployment/test-module/tp-test-ts"]
}

output "check_cm" {
	value = module.test.kustomization.manifests["_/ConfigMap/test-module/tp-ocm-ts"]
}

output "check_s" {
	value = module.test.kustomization.manifests["_/Secret/test-module/tp-os-ts-46f8b28mk5"]
}
`, modulePath, modulePath)
}

//
//
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
		target = {
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
		target = {
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
					resource.TestCheckOutput("check", "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"annotations\":{\"test-annotation\":\"true\"},\"name\":\"redis\"}}"),
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

	kustomize_options = {
		enable_helm = true
		helm_path = "helm"
	}
}

output "check" {
	value = data.kustomization_overlay.test.manifests["_/Service/_/redis"]
}
`
}
