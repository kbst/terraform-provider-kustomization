package kustomize

import (
	"context"
	"fmt"

	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sunstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"

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

func getOriginalModifiedCurrent(originalJSON string, modifiedJSON string, currentAllowNotFound bool, m interface{}) (original []byte, modified []byte, current []byte, err error) {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	n, err := parseJSON(modifiedJSON)
	if err != nil {
		return nil, nil, nil, err
	}
	o, err := parseJSON(originalJSON)
	if err != nil {
		return nil, nil, nil, err
	}

	setLastAppliedConfig(o, originalJSON)
	setLastAppliedConfig(n, modifiedJSON)

	gvr, err := cgvk.getGVR(o.GroupVersionKind(), false)
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

func parseJSON(json string) (ur *k8sunstructured.Unstructured, err error) {
	body := []byte(json)
	u, err := k8sruntime.Decode(k8sunstructured.UnstructuredJSONScheme, body)
	if err != nil {
		return ur, err
	}

	ur = u.(*k8sunstructured.Unstructured)

	return ur, nil
}
