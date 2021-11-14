package kustomize

import (
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/kustomize/api/resmap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/kubectl/pkg/scheme"
)

func flattenKustomizationIDs(rm resmap.ResMap, legacy bool) (ids []string, idsPrio [][]string, err error) {
	p0 := []string{}
	p1 := []string{}
	p2 := []string{}
	for _, id := range rm.AllIds() {
		kustomizationId := id.String()
		kr, err := parseKustomizationId(kustomizationId)
		if err != nil {
			return nil, nil, err
		}
		providerId := kustomizationId
		if !legacy {
			providerId = kr.toString()
		}
		ids = append(ids, providerId)

		p := determinePrefix(kr)
		if p < 5 {
			p0 = append(p0, providerId)
		} else if p == 9 {
			p2 = append(p2, providerId)
		} else {
			p1 = append(p1, providerId)
		}
	}

	idsPrio = append(idsPrio, p0)
	idsPrio = append(idsPrio, p1)
	idsPrio = append(idsPrio, p2)

	return ids, idsPrio, nil
}

func flattenKustomizationResources(rm resmap.ResMap, legacy bool) (res map[string]string, err error) {
	res = make(map[string]string)
	for _, r := range rm.Resources() {
		id, err := convertKustomizeToTerraform(r.CurId().String(), legacy)
		if err != nil {
			return nil, err
		}
		json, err := r.MarshalJSON()
		if err != nil {
			return nil, err
		}
		res[id] = string(json)
	}
	return res, nil
}

func flattenApiResponse(gvk schema.GroupVersionKind, re []byte, ma []byte) (json string, err error) {
	p, pt, err := getPatch(gvk, re, ma, ma)
	if err != nil {
		return json, fmt.Errorf("determining patch failed: %s", err)
	}

	var out []byte
	switch pt {
	case types.MergePatchType:
		out, err = jsonpatch.MergePatch(re, p)
		if err != nil {
			return json, fmt.Errorf("merge patch failed: %s", err)
		}
	case types.StrategicMergePatchType:
		versionedObject, err := scheme.Scheme.New(gvk)
		if err != nil {
			// if getPatch returns this patchType it already handled all error cases
			return json, fmt.Errorf("unexpected error creating versionedObject: %s", err)
		}

		out, err = strategicpatch.StrategicMergePatch(re, p, versionedObject)
		if err != nil {
			return json, fmt.Errorf("strategic merge patch failed: %s", err)
		}
	}

	json = string(out)

	return json, nil
}
