package kustomize

import (
	"sigs.k8s.io/kustomize/api/resmap"
)

func flattenKustomizationIDs(rm resmap.ResMap) (ids []string, nsNotSet []string, nsSet []string) {
	for _, id := range rm.AllIds() {
		ids = append(ids, id.String())

		// do not verify if a resource requires a namespace
		// only provide two additional lists of IDs
		//
		if id.Namespace == "" {
			nsNotSet = append(nsNotSet, id.String())
		} else {
			nsSet = append(nsSet, id.String())
		}
	}

	return ids, nsNotSet, nsSet
}

func flattenKustomizationResources(rm resmap.ResMap) (res map[string]string, err error) {
	res = make(map[string]string)
	for _, r := range rm.Resources() {
		id := r.CurId().String()
		json, err := r.MarshalJSON()
		if err != nil {
			return nil, err
		}
		res[id] = string(json)
	}

	return res, nil
}
