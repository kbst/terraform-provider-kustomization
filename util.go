package main

import (
	"fmt"

	k8scorev1 "k8s.io/api/core/v1"
	k8sunstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

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
