package kustomize

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"

	k8scorev1 "k8s.io/api/core/v1"
	k8svalidation "k8s.io/apimachinery/pkg/api/validation"
	k8sunstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"

	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/kubectl/pkg/scheme"
)

const lastAppliedConfigAnnotation = k8scorev1.LastAppliedConfigAnnotation
const gzipLastAppliedConfigAnnotation = "kustomization.kubestack.com/last-applied-config-gzip"

func setLastAppliedConfig(km *kManifest, gzipLastAppliedConfig bool) {
	annotations := km.resource.GetAnnotations()
	if len(annotations) == 0 {
		annotations = make(map[string]string)
	}

	annotations[lastAppliedConfigAnnotation] = string(km.json)

	if gzipLastAppliedConfig {
		needsGzip := false
		sErr := k8svalidation.ValidateAnnotationsSize(annotations)
		if sErr != nil {
			needsGzip = true
		}

		if needsGzip {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)

			_, err1 := zw.Write([]byte(km.json))

			err2 := zw.Close()

			if err1 == nil && err2 == nil {
				annotations[gzipLastAppliedConfigAnnotation] = base64.StdEncoding.EncodeToString(buf.Bytes())
				delete(annotations, lastAppliedConfigAnnotation)
			}
		}
	}

	km.resource.SetAnnotations(annotations)
	km.json, _ = km.resource.MarshalJSON()
}

func extractLastAppliedConfig(u *k8sunstructured.Unstructured, ex metav1ac.UnstructuredExtractor, gzipLastAppliedConfig bool) (lac string) {
	ac, err := ex.Extract(u, fieldManager)
	if err == nil {
		ans := ac.GetAnnotations()
		delete(ans, lastAppliedConfigAnnotation)
		delete(ans, gzipLastAppliedConfigAnnotation)
		ac.SetAnnotations(ans)
		if len(ans) == 0 {
			ac.SetAnnotations(nil)
		}

		b, err := ac.MarshalJSON()
		if err != nil {
			log.Fatal(err)
		}

		return strings.TrimRight(string(b), "\r\n")
	}

	// fall back to annotation if extract failed
	return getLastAppliedConfig(u, gzipLastAppliedConfig)
}

func getLastAppliedConfig(u *k8sunstructured.Unstructured, gzipLastAppliedConfig bool) (lac string) {
	annotations := u.GetAnnotations()

	lac, ok := annotations[lastAppliedConfigAnnotation]
	if !ok {
		lac = "{}"
	}

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

func getPatch(gvk k8sschema.GroupVersionKind, original []byte, modified []byte, current []byte) (pt k8stypes.PatchType, p []byte, err error) {
	versionedObject, err := scheme.Scheme.New(gvk)
	switch {
	case k8sruntime.IsNotRegisteredError(err):
		pt = k8stypes.MergePatchType

		preconditions := []mergepatch.PreconditionFunc{
			mergepatch.RequireKeyUnchanged("kind"),
			mergepatch.RequireMetadataKeyUnchanged("name"),
		}

		p, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
		if err != nil {
			return pt, p, fmt.Errorf("getPatch failed: %s", err)
		}
	case err != nil:
		return pt, p, fmt.Errorf("getPatch failed: %s", err)
	default:
		pt = k8stypes.StrategicMergePatchType

		lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(versionedObject)
		if err != nil {
			return pt, p, fmt.Errorf("getPatch failed: %s", err)
		}

		p, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true)
		if err != nil {
			return pt, p, fmt.Errorf("getPatch failed: %s", err)
		}
	}

	return pt, p, nil
}

// log error including caller name
func logError(m error) error {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	return fmt.Errorf("%s: %s", fn.Name(), m)
}
