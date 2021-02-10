package kustomize

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	k8sunstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//
//
// Basic acceptance test
func TestAccDataSourceKustomizationOverlay_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		//PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKustomizationOverlayConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.kustomization_overlay.test", "id"),
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
	resources = [
		"../test_kustomizations/basic/initial",
	]
}
`
}

//
//
// Test namespace attr
func TestKustomizationNamespace(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationNamespaceConfig(),
				Check: testKustomizationLoopManifests(
					"data.kustomization_overlay.test",
					"test-namespace",
					loopCheckNamespace,
				),
			},
		},
	})
}

func testKustomizationNamespaceConfig() string {
	return `
data "kustomization_overlay" "test" {
	namespace = "test-namespace"

	resources = [
		"../test_kustomizations/basic/initial",
	]
}
`
}

func loopCheckNamespace(u *k8sunstructured.Unstructured, exp interface{}) error {
	ns := u.GetNamespace()
	ens := exp.(string)

	// resources that are not namespaced
	// have namespace set to empty string
	if ns != "" && ns != ens {
		return fmt.Errorf("'namespace %q does not match %q: %q'", ns, ens, u.GroupVersionKind())
	}

	return nil
}

//
//
// Test common_annotations attr
func TestKustomizationCommonAnnotations(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKustomizationCommonAnnotationsConfig(),
				Check: testKustomizationLoopManifests(
					"data.kustomization_overlay.test",
					map[string]string{"test-annotation": "true"},
					loopCheckCommonAnnotations,
				),
			},
		},
	})
}

func testKustomizationCommonAnnotationsConfig() string {
	return `
data "kustomization_overlay" "test" {
	common_annotations = {
		test-annotation: true
	}

	resources = [
		"../test_kustomizations/basic/initial",
	]
}
`
}

func loopCheckCommonAnnotations(u *k8sunstructured.Unstructured, exp interface{}) error {
	as := u.GetAnnotations()
	ea := exp.(map[string]string)

	for k, v := range ea {
		if as[k] != v {
			return fmt.Errorf("'annotation %q does not equal %q: %q'", k, v, u.GroupVersionKind())
		}
	}

	return nil
}

//
//
// Test functions
func getDataSourceManifestsFromTestState(s *terraform.State, n string) (urs []*k8sunstructured.Unstructured, err error) {
	rs, ok := s.DeepCopy().RootModule().Resources[n]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", n)
	}

	d := dataSourceKustomization().Data(rs.Primary)

	for _, srcJSON := range d.Get("manifests").(map[string]interface{}) {
		u, err := parseJSON(srcJSON.(string))
		if err != nil {
			return nil, err
		}
		urs = append(urs, u)
	}

	return urs, nil
}

func testKustomizationLoopManifests(n string, exp interface{}, f loopCheckFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		urs, err := getDataSourceManifestsFromTestState(s, n)
		if err != nil {
			return err
		}

		var errs []error
		for _, u := range urs {
			e := f(u, exp)
			if e != nil {
				errs = append(errs, e)
			}
		}

		if len(errs) > 0 {
			return fmt.Errorf("errors: %+v", errs)
		}

		return nil
	}
}

type loopCheckFunc func(u *k8sunstructured.Unstructured, exp interface{}) error
