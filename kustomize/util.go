package kustomize

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sunstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/kubectl/pkg/scheme"
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
	// ensure that lastAppliedConfig definitely ends in one and only one newline
	return fmt.Sprintf("%s\n", strings.TrimRight(u.GetAnnotations()[lastAppliedConfig], "\r\n"))
}

func getOriginalModifiedCurrent(originalJSON string, modifiedJSON string, currentAllowNotFound bool, m interface{}) (original []byte, modified []byte, current []byte, err error) {
	client := m.(*Config).Client
	mapper := m.(*Config).Mapper

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

	mapping, err := mapper.RESTMapping(o.GroupVersionKind().GroupKind(), o.GroupVersionKind().Version)
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
		Resource(mapping.Resource).
		Namespace(namespace).
		Get(context.TODO(), name, k8smetav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) && currentAllowNotFound {
			return original, modified, current, nil
		}

		return nil, nil, nil, fmt.Errorf("reading '%s' failed: %s", mapping.Resource, err)
	}

	current, err = c.MarshalJSON()
	if err != nil {
		return nil, nil, nil, err
	}

	return original, modified, current, nil
}

func getPatch(gvk k8sschema.GroupVersionKind, original []byte, modified []byte, current []byte) (patch []byte, patchType k8stypes.PatchType, err error) {
	versionedObject, err := scheme.Scheme.New(gvk)
	switch {
	case k8sruntime.IsNotRegisteredError(err):
		patchType = k8stypes.MergePatchType

		preconditions := []mergepatch.PreconditionFunc{
			mergepatch.RequireKeyUnchanged("apiVersion"),
			mergepatch.RequireKeyUnchanged("kind"),
			mergepatch.RequireMetadataKeyUnchanged("name"),
		}

		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
		if err != nil {
			return nil, patchType, fmt.Errorf("getPatch failed: %s", err)
		}
	case err != nil:
		return nil, patchType, fmt.Errorf("getPatch failed: %s", err)
	case err == nil:
		patchType = k8stypes.StrategicMergePatchType

		lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(versionedObject)
		if err != nil {
			return nil, patchType, fmt.Errorf("getPatch failed: %s", err)
		}

		patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true)
		if err != nil {
			return nil, patchType, fmt.Errorf("getPatch failed: %s", err)
		}
	}

	return patch, patchType, nil
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

// log error including caller name and k8s resource
func logErrorForResource(u *k8sunstructured.Unstructured, m error) error {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	return fmt.Errorf(
		"%s: apiVersion: %q, kind: %q, namespace: %q name: %q: %s",
		fn.Name(),
		u.GroupVersionKind().GroupVersion(),
		u.GroupVersionKind().Kind,
		u.GetNamespace(),
		u.GetName(),
		m)
}

// log error including caller name
func logError(m error) error {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	return fmt.Errorf(
		"%s: %s",
		fn.Name(),
		m)
}
