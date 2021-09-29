package kustomize

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sunstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
)

//
//
// Basic test
func TestAccResourceKustomization_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config with a svc and deployment in a namespace
			{
				Config: testAccResourceKustomizationConfig_basicInitial("test_kustomizations/basic/initial"),
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
				Config: testAccResourceKustomizationConfig_basicModified("test_kustomizations/basic/modified"),
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
				Config: testAccResourceKustomizationConfig_basicInitial("test_kustomizations/basic/initial"),
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

func testAccResourceKustomizationConfig_basicInitial(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
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

func testAccResourceKustomizationConfig_basicModified(path string) string {
	return testAccResourceKustomizationConfig_basicInitial(path) + `
resource "kustomization_resource" "dep2" {
	manifest = data.kustomization_build.test.manifests["apps_v1_Deployment|test-basic|test2"]
}
`
}

//
//
// Import test invalid id
func TestAccResourceKustomization_importInvalidID(t *testing.T) {

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
func TestAccResourceKustomization_updateInplace(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config with a svc and deployment in a namespace
			{
				Config: testAccResourceKustomizationConfig_updateInplace("test_kustomizations/update_inplace/initial"),
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
				Config: testAccResourceKustomizationConfig_updateInplace("test_kustomizations/update_inplace/modified"),
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
				Config: testAccResourceKustomizationConfig_updateInplace("test_kustomizations/update_inplace/initial"),
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

func testAccResourceKustomizationConfig_updateInplace(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
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
// Update_Merge_Fallback Test
func TestAccResourceKustomization_updateMergeFallback(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config with env set by a value
			{
				Config: testAccResourceKustomizationConfig_updateMergeFallbackInitial("test_kustomizations/update_merge_fallback/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep", "id"),
				),
			},
			//
			//
			// Applying modified config replacing env value with valueFrom
			{
				Config: testAccResourceKustomizationConfig_updateMergeFallbackModified("test_kustomizations/update_merge_fallback/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.svc", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.dep", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.cm", "id"),
				),
			},
		},
	})
}

func testAccResourceKustomizationConfig_updateMergeFallbackInitial(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
resource "kustomization_resource" "ns" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Namespace|~X|test-update-merge-fallback"]
}

resource "kustomization_resource" "svc" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Service|test-update-merge-fallback|test"]
}

resource "kustomization_resource" "dep" {
	manifest = data.kustomization_build.test.manifests["apps_v1_Deployment|test-update-merge-fallback|test"]
}
`
}

func testAccResourceKustomizationConfig_updateMergeFallbackModified(path string) string {
	return testAccResourceKustomizationConfig_updateMergeFallbackInitial(path) + `
resource "kustomization_resource" "cm" {
	manifest = data.kustomization_build.test.manifests["~G_v1_ConfigMap|test-update-merge-fallback|test-envfrom-kmdg664296"]
}
`
}

//
//
// Update_Recreate Test
func TestAccResourceKustomization_updateRecreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config with a svc and deployment in a namespace
			{
				Config: testAccResourceKustomizationConfig_updateRecreate("test_kustomizations/update_recreate/initial"),
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
				Config: testAccResourceKustomizationConfig_updateRecreate("test_kustomizations/update_recreate/modified"),
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
				Config: testAccResourceKustomizationConfig_updateRecreate("test_kustomizations/update_recreate/initial"),
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

func testAccResourceKustomizationConfig_updateRecreate(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
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
func TestAccResourceKustomization_updateRecreateNameOrNamespaceChange(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config with a svc and deployment in a namespace
			{
				Config: testAccResourceKustomizationConfig_updateRecreateNameOrNamespaceChange("test_kustomizations/update_recreate_name_or_namespace_change/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.cm", "id"),
				),
			},
			//
			//
			// Applying modified config changing the immutable label selectors
			{
				Config: testAccResourceKustomizationConfig_updateRecreateNameOrNamespaceChangeModified("test_kustomizations/update_recreate_name_or_namespace_change/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.cm", "id"),
				),
			},
		},
	})
}

func testAccResourceKustomizationConfig_updateRecreateNameOrNamespaceChange(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
resource "kustomization_resource" "ns" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Namespace|~X|test-update-recreate-name-or-namespace-change"]
}

resource "kustomization_resource" "cm" {
	manifest = data.kustomization_build.test.manifests["~G_v1_ConfigMap|test-update-recreate-name-or-namespace-change|test"]
}
`
}

func testAccResourceKustomizationConfig_updateRecreateNameOrNamespaceChangeModified(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
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
func TestAccResourceKustomization_updateRecreateStatefulSet(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial statefulset
			{
				Config: testAccResourceKustomizationConfig_updateRecreateStatefulSet("test_kustomizations/update_recreate_statefulset/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.ss", "id"),
				),
			},
			//
			//
			// Applying modified statefulset that requires a destroy and recreate
			{
				Config: testAccResourceKustomizationConfig_updateRecreateStatefulSet("test_kustomizations/update_recreate_statefulset/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kustomization_resource.ns", "id"),
					resource.TestCheckResourceAttrSet("kustomization_resource.ss", "id"),
				),
			},
		},
	})
}

func testAccResourceKustomizationConfig_updateRecreateStatefulSet(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
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
// CRD Test
func TestAccResourceKustomization_crd(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying both namespaced and cluster wide CRD
			// and one custom object of each CRD
			{
				Config: testAccResourceKustomizationConfig_crd("test_kustomizations/crd/initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.clusteredcrd",
						"id"),
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.namespacedcrd",
						"id"),
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.clusteredco",
						"id"),
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.namespacedco",
						"id"),
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.ns",
						"id"),
				),
			},
			//
			//
			// Modify each CO's spec with a patch
			{
				Config: testAccResourceKustomizationConfig_crd("test_kustomizations/crd/modified"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.clusteredcrd",
						"id"),
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.namespacedcrd",
						"id"),
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.clusteredco",
						"id"),
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.namespacedco",
						"id"),
					resource.TestCheckResourceAttrSet(
						"kustomization_resource.ns",
						"id"),
				),
			},
			//
			//
			// Test state import
			{
				ResourceName:      "kustomization_resource.test[\"apiextensions.k8s.io_v1_CustomResourceDefinition|~X|clusteredcrds.test.example.com\"]",
				ImportStateId:     "apiextensions.k8s.io_v1_CustomResourceDefinition|~X|clusteredcrds.test.example.com",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceKustomizationConfig_crd(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
resource "kustomization_resource" "clusteredcrd" {
	manifest = data.kustomization_build.test.manifests["apiextensions.k8s.io_v1_CustomResourceDefinition|~X|clusteredcrds.test.example.com"]
}

resource "kustomization_resource" "namespacedcrd" {
	manifest = data.kustomization_build.test.manifests["apiextensions.k8s.io_v1_CustomResourceDefinition|~X|namespacedcrds.test.example.com"]
}

resource "kustomization_resource" "clusteredco" {
	manifest = data.kustomization_build.test.manifests["test.example.com_v1alpha1_Clusteredcrd|~X|clusteredco"]
}

resource "kustomization_resource" "namespacedco" {
	manifest = data.kustomization_build.test.manifests["test.example.com_v1alpha1_Namespacedcrd|test-crd|namespacedco"]
}

resource "kustomization_resource" "ns" {
	manifest = data.kustomization_build.test.manifests["~G_v1_Namespace|~X|test-crd"]
}
`
}

//
//
// Webhook Test
func TestAccResourceKustomization_webhook(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Creating initial webhook
			{
				Config: testAccResourceKustomizationConfig_webhook("test_kustomizations/webhook/initial"),
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
				Config: testAccResourceKustomizationConfig_webhook("test_kustomizations/webhook/modified"),
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

func testAccResourceKustomizationConfig_webhook(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
resource "kustomization_resource" "webhook" {
	manifest = data.kustomization_build.test.manifests["admissionregistration.k8s.io_v1_ValidatingWebhookConfiguration|~X|pod-policy.example.com"]
}
`
}

//
//
// TransformerConfigs test
func TestAccResourceKustomization_transformerConfigs(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//
			//
			// Applying initial config without the test label
			{
				Config: testAccResourceKustomizationConfig_transformerConfigs("test_kustomizations/transformer_configs/initial"),
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
				Config: testAccResourceKustomizationConfig_transformerConfigs("test_kustomizations/transformer_configs/modified"),
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

func testAccResourceKustomizationConfig_transformerConfigs(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
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

//
//
// Test check functions

func testAccCheckDeploymentPurged(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Config).Client

		gvr := k8sschema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}
		namespace := "test"
		name := "test2"

		_, k8serr := client.
			Resource(gvr).
			Namespace(namespace).
			Get(context.TODO(), name, k8smetav1.GetOptions{})
		if k8serr != nil {
			if !k8serrors.IsNotFound(k8serr) {
				return fmt.Errorf("Unexpected error from K8s api: %s", k8serr)
			}
		} else {
			return fmt.Errorf("Resource not purged from K8s api: %s", n)
		}

		return nil
	}
}

func getResourceFromTestState(s *terraform.State, n string) (ur *k8sunstructured.Unstructured, err error) {
	rs, ok := s.RootModule().Resources[n]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", n)
	}

	srcJSON := rs.Primary.Attributes["manifest"]
	u, err := parseJSON(srcJSON)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func getResourceFromK8sAPI(u *k8sunstructured.Unstructured) (resp *k8sunstructured.Unstructured, err error) {
	client := testAccProvider.Meta().(*Config).Client
	mapper := testAccProvider.Meta().(*Config).Mapper

	mapping, err := mapper.RESTMapping(u.GroupVersionKind().GroupKind(), u.GroupVersionKind().Version)
	if err != nil {
		return nil, err
	}
	namespace := u.GetNamespace()
	name := u.GetName()

	resp, err = client.
		Resource(mapping.Resource).
		Namespace(namespace).
		Get(context.TODO(), name, k8smetav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func testAccCheckManifestAnnotation(n string, k string, v string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		u, err := getResourceFromTestState(s, n)
		if err != nil {
			return err
		}

		resp, err := getResourceFromK8sAPI(u)
		if err != nil {
			return err
		}

		annotations := resp.GetAnnotations()
		a, ok := annotations[k]
		if !ok {
			return fmt.Errorf("Annotation missing: %s", k)
		}

		if a != v {
			return fmt.Errorf("Annotation value incorrect: expected %s, got %s", v, a)
		}

		return nil
	}
}

func testAccCheckManifestAnnotationAbsent(n string, k string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		u, err := getResourceFromTestState(s, n)
		if err != nil {
			return err
		}

		resp, err := getResourceFromK8sAPI(u)
		if err != nil {
			return err
		}

		annotations := resp.GetAnnotations()
		_, ok := annotations[k]
		if ok {
			return fmt.Errorf("Unexpected annotation exists: %s", k)
		}

		return nil
	}
}

func testAccCheckManifestLabel(n string, k string, v string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		u, err := getResourceFromTestState(s, n)
		if err != nil {
			return err
		}

		resp, err := getResourceFromK8sAPI(u)
		if err != nil {
			return err
		}

		labels := resp.GetLabels()
		a, ok := labels[k]
		if !ok {
			return fmt.Errorf("Label missing: %s", k)
		}

		if a != v {
			return fmt.Errorf("Label value incorrect: expected %s, got %s", v, a)
		}

		return nil
	}
}

func testAccCheckManifestLabelAbsent(n string, k string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		u, err := getResourceFromTestState(s, n)
		if err != nil {
			return err
		}

		resp, err := getResourceFromK8sAPI(u)
		if err != nil {
			return err
		}

		labels := resp.GetLabels()
		_, ok := labels[k]
		if ok {
			return fmt.Errorf("Unexpected label exists: %s", k)
		}

		return nil
	}
}

func testAccCheckManifestSelector(n string, k string, v string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		u, err := getResourceFromTestState(s, n)
		if err != nil {
			return err
		}

		resp, err := getResourceFromK8sAPI(u)
		if err != nil {
			return err
		}

		matchLabels, ok, err := k8sunstructured.NestedStringMap(resp.Object, "spec", "selector", "matchLabels")
		if !ok {
			return fmt.Errorf("Selector matchLabels missing from spec")
		}
		if err != nil {
			return err
		}

		a, ok := matchLabels[k]
		if !ok {
			return fmt.Errorf("Selector matchLabels missing: %s, %v", k, matchLabels)
		}

		if a != v {
			return fmt.Errorf("Selector matchLabels value incorrect: expected %s, got %s", v, a)
		}

		return nil
	}
}

func testAccCheckManifestSelectorAbsent(n string, k string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		u, err := getResourceFromTestState(s, n)
		if err != nil {
			return err
		}

		resp, err := getResourceFromK8sAPI(u)
		if err != nil {
			return err
		}

		matchLabels, ok, err := k8sunstructured.NestedStringMap(resp.Object, "spec", "selector", "matchLabels")
		if !ok {
			return fmt.Errorf("Selector matchLabels missing from spec")
		}
		if err != nil {
			return err
		}

		_, ok = matchLabels[k]
		if ok {
			return fmt.Errorf("Unexpected selector matchLabels: %s", k)
		}

		return nil
	}
}
