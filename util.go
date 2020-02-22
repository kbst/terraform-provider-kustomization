package main

import (
	"fmt"

	k8scorev1 "k8s.io/api/core/v1"
	k8sunstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"

	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
)

const lastAppliedConfig = k8scorev1.LastAppliedConfigAnnotation

func setLastAppliedConfig(u *k8sunstructured.Unstructured, srcJSON string) {
	annotations := u.GetAnnotations()
	if len(annotations) == 0 {
		annotations = make(map[string]string)
	}
	annotations[lastAppliedConfig] = srcJSON
	u.SetAnnotations(annotations)
}

func getLastAppliedConfig(u *k8sunstructured.Unstructured) string {
	return u.GetAnnotations()[lastAppliedConfig]
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
