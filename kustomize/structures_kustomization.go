package kustomize

import (
	"sigs.k8s.io/kustomize/api/resmap"
)

func flattenKustomizationIDs(rm resmap.ResMap) (ids []string, idsPrio [][]string, err error) {
	p0 := []string{}
	p1 := []string{}
	p2 := []string{}
	for _, id := range rm.AllIds() {
		kr := &kManifestId{
			group:     id.Group,
			kind:      id.Kind,
			namespace: id.Namespace,
			name:      id.Name,
		}

		ids = append(ids, kr.toString())

		p := determinePrefix(kr)
		if p < 5 {
			p0 = append(p0, kr.toString())
		} else if p == 9 {
			p2 = append(p2, kr.toString())
		} else {
			p1 = append(p1, kr.toString())
		}
	}

	idsPrio = append(idsPrio, p0)
	idsPrio = append(idsPrio, p1)
	idsPrio = append(idsPrio, p2)

	return ids, idsPrio, nil
}

func flattenKustomizationResources(rm resmap.ResMap) (res map[string]string, err error) {
	res = make(map[string]string)
	for _, r := range rm.Resources() {
		kr := &kManifestId{
			group:     r.CurId().Group,
			kind:      r.CurId().Kind,
			namespace: r.GetNamespace(),
			name:      r.GetName(),
		}

		json, err := r.MarshalJSON()
		if err != nil {
			return nil, err
		}
		res[kr.toString()] = string(json)
	}
	return res, nil
}
