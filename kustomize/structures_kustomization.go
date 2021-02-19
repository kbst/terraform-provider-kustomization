package kustomize

import (
	"sigs.k8s.io/kustomize/api/resmap"
)

func flattenKustomizationIDs(rm resmap.ResMap) (ids []string, idsPrio [][]string) {
	p0 := []string{}
	p1 := []string{}
	p2 := []string{}
	for _, id := range rm.AllIds() {
		ids = append(ids, id.String())

		p := determinePrefix(id.Gvk)
		if p < 5 {
			p0 = append(p0, id.String())
		} else if p == 9 {
			p2 = append(p2, id.String())
		} else {
			p1 = append(p1, id.String())
		}
	}

	idsPrio = append(idsPrio, p0)
	idsPrio = append(idsPrio, p1)
	idsPrio = append(idsPrio, p2)

	return ids, idsPrio
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
