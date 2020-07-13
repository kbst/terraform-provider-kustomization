package kustomize

import (
	"sigs.k8s.io/kustomize/api/resmap"
	"strings"
)

func flattenKustomizationIDs(rm resmap.ResMap) (ids []string) {
	for _, id := range rm.AllIds() {
		ids = append(ids, id.String())
	}

	return ids
}

func flattenKustomizationResources(rm resmap.ResMap) (res map[string]string, err error) {
	res = make(map[string]string)
	for _, r := range rm.Resources() {
		id := r.CurId().String()
		json, err := r.MarshalJSON()
		if err != nil {
			return nil, err
		}
		res[id] = strings.Trim(string(json), "\n")
	}

	return res, nil
}
