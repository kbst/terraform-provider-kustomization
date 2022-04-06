package kustomize

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"

	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8svalidation "k8s.io/apimachinery/pkg/api/validation"
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

const lastAppliedConfigAnnotation = k8scorev1.LastAppliedConfigAnnotation
const gzipLastAppliedConfigAnnotation = "kustomization.kubestack.com/last-applied-config-gzip"

func setLastAppliedConfig(u *k8sunstructured.Unstructured, srcJSON string, gzipLastAppliedConfig bool) {
	annotations := u.GetAnnotations()
	if len(annotations) == 0 {
		annotations = make(map[string]string)
	}

	annotations[lastAppliedConfigAnnotation] = srcJSON

	if gzipLastAppliedConfig {
		needsGzip := false
		sErr := k8svalidation.ValidateAnnotationsSize(annotations)
		if sErr != nil {
			needsGzip = true
		}
	
		if needsGzip {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
	
			_, err1 := zw.Write([]byte(srcJSON))
	
			err2 := zw.Close()
	
			if err1 == nil && err2 == nil {
				annotations[gzipLastAppliedConfigAnnotation] = base64.StdEncoding.EncodeToString(buf.Bytes())
				delete(annotations, lastAppliedConfigAnnotation)
			}
		}
	}

	u.SetAnnotations(annotations)
}

func getLastAppliedConfig(u *k8sunstructured.Unstructured, gzipLastAppliedConfig bool) (lac string) {
	annotations := u.GetAnnotations()

	lac = u.GetAnnotations()[lastAppliedConfigAnnotation]

	if gzipLastAppliedConfig {
		// read the compressed lac if available
		if gzEnc, ok := annotations[gzipLastAppliedConfigAnnotation]; ok {
			gzDec, err := base64.StdEncoding.DecodeString(gzEnc)
			if err != nil {
				log.Fatal(err)
			}
	
			var buf bytes.Buffer
			buf.Write(gzDec)
	
			zr, err1 := gzip.NewReader(&buf)
	
			lacBuf := new(strings.Builder)
			_, err2 := io.Copy(lacBuf, zr)
	
			err3 := zr.Close()
	
			// in case of any error, fall back to the uncompressed lac
			if err1 == nil && err2 == nil && err3 == nil {
				lac = lacBuf.String()
			}
		}
	}

	return strings.TrimRight(lac, "\r\n")
}

func getOriginalModifiedCurrent(originalJSON string, modifiedJSON string, currentAllowNotFound bool, m interface{}) (original []byte, modified []byte, current []byte, err error) {
	client := m.(*Config).Client
	mapper := m.(*Config).Mapper
	gzipLastAppliedConfig := m.(*Config).GzipLastAppliedConfig

	n, err := parseJSON(modifiedJSON)
	if err != nil {
		return nil, nil, nil, err
	}
	o, err := parseJSON(originalJSON)
	if err != nil {
		return nil, nil, nil, err
	}

	setLastAppliedConfig(o, originalJSON, gzipLastAppliedConfig)
	setLastAppliedConfig(n, modifiedJSON, gzipLastAppliedConfig)

	mapping, err := mapper.RESTMapping(n.GroupVersionKind().GroupKind(), n.GroupVersionKind().Version)
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
