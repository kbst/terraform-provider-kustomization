package kustomize

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

//
//
// Basic test
func TestAccResourceKustomization_legacy_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config with a svc and deployment in a namespace
			{
				Config: testAccResourceKustomizationConfig_legacy_basicInitial("test_kustomizations/basic/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
				),
			},
			//
			//
			// Applying modified config adding another deployment to the namespace
			{
				Config: testAccResourceKustomizationConfig_legacy_basicModified("test_kustomizations/basic/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep2", "id"),
				),
			},
			//
			//
			// Reverting back to initial config with only one deployment
			// check that second deployment was purged
			{
				Config: testAccResourceKustomizationConfig_legacy_basicInitial("test_kustomizations/basic/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
					testAccCheckDeploymentPurged("kustomization_resource.dep2"),
				),
			},
			//
			//
			// Test state import
			{
				ResourceName:      "kustomization_resource.test[\"~G_v1_Namespace|~X|test-basic\"]",
				ImportStateId:     "~G_v1_Namespace|~X|test-basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceKustomizationConfig_legacy_basicInitial(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path, true) + `
resource "kustomization_resource" "ns" {
       manifest = data.kustomization_build.test.manifests["~G_v1_Namespace|~X|test-basic"]
}

resource "kustomization_resource" "svc" {
       manifest = data.kustomization_build.test.manifests["~G_v1_Service|test-basic|test"]
 }

resource "kustomization_resource" "dep1" {
       manifest = data.kustomization_build.test.manifests["apps_v1_Deployment|test-basic|test"]
}
`
}

func testAccResourceKustomizationConfig_legacy_basicModified(path string) string {
	return testAccResourceKustomizationConfig_legacy_basicInitial(path) + `
resource "kustomization_resource" "dep2" {
       manifest = data.kustomization_build.test.manifests["apps_v1_Deployment|test-basic|test2"]
}
`
}

//
//
// Import test invalid id
func TestAccResourceKustomization_importLegacyInvalidID(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Test state import
			{
				ResourceName:      "kustomization_resource.test[\"~G_v1_Namespace|~X|test-basic\"]",
				ImportStateId:     "invalidID",
				ImportState:       true,
				ImportStateVerify: true,
				ExpectError:       regexp.MustCompile("invalid ID: \"invalidID\", valid IDs look like: \"~G_v1_Namespace|~X|example\""),
			},
		},
	})
}

//
//
// Update_Inplace Test
func TestAccResourceKustomization_legacy_updateInplace(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config with a svc and deployment in a namespace
			{
				Config: testAccResourceKustomizationConfig_legacy_updateInplace("test_kustomizations/update_inplace/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
				),
			},
			//
			//
			// Applying modified config adding an annotation to each resource
			{
				Config: testAccResourceKustomizationConfig_legacy_updateInplace("test_kustomizations/update_inplace/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
					testAccCheckManifestAnnotation("kustomization_resource.ns", "test_annotation", "added"),
					testAccCheckManifestAnnotation("kustomization_resource.svc", "test_annotation", "added"),
					testAccCheckManifestAnnotation("kustomization_resource.dep1", "test_annotation", "added"),
				),
			},
			//
			//
			// Applying initial config again, ensure annotations are removed again
			{
				Config: testAccResourceKustomizationConfig_legacy_updateInplace("test_kustomizations/update_inplace/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
					testAccCheckManifestAnnotationAbsent("kustomization_resource.ns", "test_annotation"),
					testAccCheckManifestAnnotationAbsent("kustomization_resource.svc", "test_annotation"),
					testAccCheckManifestAnnotationAbsent("kustomization_resource.dep1", "test_annotation"),
				),
			},
			//
			//
			// Test state import
			{
				ResourceName:      "kustomization_resource.test[\"~G_v1_Namespace|~X|test-update-inplace\"]",
				ImportStateId:     "~G_v1_Namespace|~X|test-update-inplace",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceKustomizationConfig_legacy_updateInplace(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path, true) + `
resource "kustomization_resource" "ns" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Namespace|~X|test-update-inplace"]
}

resource "kustomization_resource" "svc" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Service|test-update-inplace|test"]
}

resource "kustomization_resource" "dep1" {
	manifest = data.kustomization_build.test.manifests["apps_v1_Deployment|test-update-inplace|test"]
}
`
}

//
//
// Update_Recreate Test
func TestAccResourceKustomization_legacy_updateRecreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config with a svc and deployment in a namespace
			{
				Config: testAccResourceKustomizationConfig_legacy_updateRecreate("test_kustomizations/update_recreate/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
				),
			},
			//
			//
			// Applying modified config changing the immutable label selectors
			{
				Config: testAccResourceKustomizationConfig_legacy_updateRecreate("test_kustomizations/update_recreate/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
					testAccCheckManifestSelector("kustomization_resource.dep1", "test-label", "added"),
				),
			},
			//
			//
			// Applying initial config again, ensure label selector is back to original state
			{
				Config: testAccResourceKustomizationConfig_legacy_updateRecreate("test_kustomizations/update_recreate/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
					testAccCheckManifestSelectorAbsent("kustomization_resource.dep1", "test-label"),
				),
			},
			//
			//
			// Test state import
			{
				ResourceName:      "kustomization_resource.test[\"~G_v1_Namespace|~X|test-update-recreate\"]",
				ImportStateId:     "~G_v1_Namespace|~X|test-update-recreate",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceKustomizationConfig_legacy_updateRecreate(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path, true) + `
resource "kustomization_resource" "ns" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Namespace|~X|test-update-recreate"]
}

resource "kustomization_resource" "svc" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Service|test-update-recreate|test"]
}

resource "kustomization_resource" "dep1" {
	manifest = data.kustomization_build.test.manifests["apps_v1_Deployment|test-update-recreate|test"]
}
`
}

//
//
// Update_Recreate_Name_Or_Namespace_Change Test
func TestAccResourceKustomization_legacy_updateRecreateNameOrNamespaceChange(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config with a svc and deployment in a namespace
			{
				Config: testAccResourceKustomizationConfig_legacy_updateRecreateNameOrNamespaceChange("test_kustomizations/update_recreate_name_or_namespace_change/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.cm", "id"),
				),
			},
			//
			//
			// Applying modified config changing the immutable label selectors
			{
				Config: testAccResourceKustomizationConfig_legacy_updateRecreateNameOrNamespaceChangeModified("test_kustomizations/update_recreate_name_or_namespace_change/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.cm", "id"),
				),
			},
		},
	})
}

func testAccResourceKustomizationConfig_legacy_updateRecreateNameOrNamespaceChange(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path, true) + `
resource "kustomization_resource" "ns" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Namespace|~X|test-update-recreate-name-or-namespace-change"]
}

resource "kustomization_resource" "cm" {
	manifest = data.kustomization_build.test.manifests["~G_v1_ConfigMap|test-update-recreate-name-or-namespace-change|test"]
}
`
}

func testAccResourceKustomizationConfig_legacy_updateRecreateNameOrNamespaceChangeModified(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path, true) + `
resource "kustomization_resource" "ns" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Namespace|~X|test-update-recreate-name-or-namespace-change-modified"]
}

resource "kustomization_resource" "cm" {
	manifest = data.kustomization_build.test.manifests["~G_v1_ConfigMap|test-update-recreate-name-or-namespace-change-modified|test"]
}
`
}

//
//
// Update_Recreate_StatefulSet Test
func TestAccResourceKustomization_legacy_updateRecreateStatefulSet(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial statefulset
			{
				Config: testAccResourceKustomizationConfig_legacy_updateRecreateStatefulSet("test_kustomizations/update_recreate_statefulset/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.ss", "id"),
				),
			},
			//
			//
			// Applying modified statefulset that requires a destroy and recreate
			{
				Config: testAccResourceKustomizationConfig_legacy_updateRecreateStatefulSet("test_kustomizations/update_recreate_statefulset/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.ss", "id"),
				),
			},
		},
	})
}

func testAccResourceKustomizationConfig_legacy_updateRecreateStatefulSet(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path, true) + `
resource "kustomization_resource" "ns" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Namespace|~X|test-update-recreate-statefulset"]
}

resource "kustomization_resource" "ss" {
	manifest = data.kustomization_build.test.manifests["apps_v1_StatefulSet|test-update-recreate-statefulset|test"]
}
`
}

//
//
// Webhook Test
func TestAccResourceKustomization_legacy_webhook(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Creating initial webhook
			{
				Config: testAccResourceKustomizationConfig_legacy_webhook("test_kustomizations/webhook/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.webhook",
						"id"),
				),
			},
			//
			//
			// Applying modified webhook
			{
				Config: testAccResourceKustomizationConfig_legacy_webhook("test_kustomizations/webhook/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.webhook",
						"id"),
				),
			},
			//
			//
			// Test state import
			{
				ResourceName:      "kustomization_resource.test[\"admissionregistration.k8s.io_v1_ValidatingWebhookConfiguration|~X|pod-policy.example.com\"]",
				ImportStateId:     "admissionregistration.k8s.io_v1_ValidatingWebhookConfiguration|~X|pod-policy.example.com",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceKustomizationConfig_legacy_webhook(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path, true) + `
resource "kustomization_resource" "webhook" {
	manifest = data.kustomization_build.test.manifests["admissionregistration.k8s.io_v1_ValidatingWebhookConfiguration|~X|pod-policy.example.com"]
}
`
}

//
//
// TransformerConfigs test
func TestAccResourceKustomization_legacy_transformerConfigs(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config without the test label
			{
				Config: testAccResourceKustomizationConfig_legacy_transformerConfigs("test_kustomizations/transformer_configs/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
					testAccCheckManifestLabelAbsent("kustomization_resource.dep1", "test.example.com/test-label"),
					testAccCheckManifestSelectorAbsent("kustomization_resource.dep1", "test.example.com/test-label"),
				),
			},
			//
			//
			// Applying modified config adding the test label
			{
				Config: testAccResourceKustomizationConfig_legacy_transformerConfigs("test_kustomizations/transformer_configs/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep1", "id"),
					testAccCheckManifestLabel("kustomization_resource.dep1", "test.example.com/test-label", "test-value"),
					testAccCheckManifestSelectorAbsent("kustomization_resource.dep1", "test.example.com/test-label"),
				),
			},
		},
	})
}

func testAccResourceKustomizationConfig_legacy_transformerConfigs(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path, true) + `
resource "kustomization_resource" "ns" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Namespace|~X|test-transformer-config"]
}

resource "kustomization_resource" "svc" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Service|test-transformer-config|test"]
}

resource "kustomization_resource" "dep1" {
	manifest = data.kustomization_build.test.manifests["apps_v1_Deployment|test-transformer-config|test"]
}
`
}
