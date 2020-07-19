package kustomize

import (
	"context"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sunstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"

	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
)

// Get a simplified json from resp by stripping excessive fields based on targeted.
// Mainly for eliminating unwanted noisy diff in terraform plan.
func getSimplified(resp *k8sunstructured.Unstructured, targeted []byte) string {
	bytes, _ := resp.MarshalJSON()
	simplified, _ := simplifyJSON(bytes, targeted)
	return string(simplified)
}

func simplifyJSON(full, targeted []byte) (simplified []byte, err error) {
	if cap(targeted) == 0 {
		return full, nil
	}
	var preconditions []mergepatch.PreconditionFunc

	patch1, err := jsonmergepatch.CreateThreeWayJSONMergePatch(targeted, full, full, preconditions...)
	if err != nil {
		return nil, err
	}

	patched, err := jsonpatch.MergePatch(full, patch1)
	if err != nil {
		return nil, err
	}

	patch2, err := jsonmergepatch.CreateThreeWayJSONMergePatch(full, targeted, targeted, preconditions...)
	if err != nil {
		return nil, err
	}

	return jsonpatch.MergePatch(patched, patch2)
}

func getOriginalModifiedCurrent(originalJSON string, modifiedJSON string, currentAllowNotFound bool, m interface{}) (original []byte, modified []byte, current []byte, err error) {
	client := m.(*Config).Client
	clientset := m.(*Config).Clientset

	n, err := parseJSON(modifiedJSON)
	if err != nil {
		return nil, nil, nil, err
	}
	o, err := parseJSON(originalJSON)
	if err != nil {
		return nil, nil, nil, err
	}

	gvr, err := getGVR(o.GroupVersionKind(), clientset)
	if err != nil {
		return nil, nil, nil, err
	}
	namespace := o.GetNamespace()
	name := o.GetName()

	original, err = o.MarshalJSON()
	if err != nil {
		return nil, nil, nil, err
	}

	modified, err = n.MarshalJSON()
	if err != nil {
		return nil, nil, nil, err
	}

	c, err := client.
		Resource(gvr).
		Namespace(namespace).
		Get(context.TODO(), name, k8smetav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) && currentAllowNotFound {
			return original, modified, current, nil
		}

		return nil, nil, nil, fmt.Errorf("reading '%s' failed: %s", gvr, err)
	}

	current, err = c.MarshalJSON()
	if err != nil {
		return nil, nil, nil, err
	}

	return original, modified, current, nil
}

func getPatch(original []byte, modified []byte, current []byte) (patch []byte, err error) {
	preconditions := []mergepatch.PreconditionFunc{
		mergepatch.RequireKeyUnchanged("apiVersion"),
		mergepatch.RequireKeyUnchanged("kind"),
		mergepatch.RequireMetadataKeyUnchanged("name")}
	patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(
		original, modified, current, preconditions...)
	if err != nil {
		return nil, fmt.Errorf("CreateThreeWayJSONMergePatch failed: %s", err)
	}
	return patch, nil
}

func getGVR(gvk k8sschema.GroupVersionKind, cs *kubernetes.Clientset) (gvr k8sschema.GroupVersionResource, err error) {
	agr, err := restmapper.GetAPIGroupResources(cs.Discovery())
	if err != nil {
		return gvr, fmt.Errorf("discovering API group resources failed: %s", err)
	}

	rm := restmapper.NewDiscoveryRESTMapper(agr)

	gk := k8sschema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil {
		return gvr, fmt.Errorf("mapping GroupKind failed for '%s': %s", gvk, err)
	}

	gvr = mapping.Resource

	return gvr, nil
}

func parseJSON(json string) (ur *k8sunstructured.Unstructured, err error) {
	body := []byte(json)
	u, err := k8sruntime.Decode(k8sunstructured.UnstructuredJSONScheme, body)
	if err != nil {
		return ur, err
	}

	ur = u.(*k8sunstructured.Unstructured)

	return ur, nil
}
