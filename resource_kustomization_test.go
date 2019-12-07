package main

import (
	"fmt"
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
		},
	})
}

func testAccResourceKustomizationConfig_basicInitial(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
resource "kustomization_resource" "ns" {
  manifest = data.kustomization.test.manifests["~G_v1_Namespace|~X|test-basic"]
}

resource "kustomization_resource" "svc" {
  manifest = data.kustomization.test.manifests["~G_v1_Service|test-basic|test"]
}

resource "kustomization_resource" "dep1" {
  manifest = data.kustomization.test.manifests["apps_v1_Deployment|test-basic|test"]
}
`
}

func testAccResourceKustomizationConfig_basicModified(path string) string {
	return testAccResourceKustomizationConfig_basicInitial(path) + `
resource "kustomization_resource" "dep2" {
  manifest = data.kustomization.test.manifests["apps_v1_Deployment|test-basic|test2"]
}
`
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
		},
	})
}

func testAccResourceKustomizationConfig_updateInplace(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
resource "kustomization_resource" "ns" {
  manifest = data.kustomization.test.manifests["~G_v1_Namespace|~X|test-update-inplace"]
}

resource "kustomization_resource" "svc" {
  manifest = data.kustomization.test.manifests["~G_v1_Service|test-update-inplace|test"]
}

resource "kustomization_resource" "dep1" {
  manifest = data.kustomization.test.manifests["apps_v1_Deployment|test-update-inplace|test"]
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
		},
	})
}

func testAccResourceKustomizationConfig_updateRecreate(path string) string {
	return testAccDataSourceKustomizationConfig_basic(path) + `
resource "kustomization_resource" "ns" {
  manifest = data.kustomization.test.manifests["~G_v1_Namespace|~X|test-update-recreate"]
}

resource "kustomization_resource" "svc" {
  manifest = data.kustomization.test.manifests["~G_v1_Service|test-update-recreate|test"]
}

resource "kustomization_resource" "dep1" {
  manifest = data.kustomization.test.manifests["apps_v1_Deployment|test-update-recreate|test"]
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
			Get(name, k8smetav1.GetOptions{})
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

	gvr := getGVR(u)
	namespace := u.GetNamespace()
	name := u.GetName()

	resp, err = client.
		Resource(gvr).
		Namespace(namespace).
		Get(name, k8smetav1.GetOptions{})
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
