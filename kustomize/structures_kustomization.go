package kustomize

import (
	"sigs.k8s.io/kustomize/api/resmap"
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
