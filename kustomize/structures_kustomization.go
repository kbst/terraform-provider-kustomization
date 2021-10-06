package kustomize

import (
	"sigs.k8s.io/kustomize/api/resmap"
)

func flattenKustomizationIDs(rm resmap.ResMap, legacy bool) (ids []string, idsPrio [][]string, err error) {
	p0 := []string{}
	p1 := []string{}
	p2 := []string{}
	for _, id := range rm.AllIds() {
		versionLessId, err := convertKustomizeToTerraform(id.String(), legacy)
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, versionLessId)

		p := determinePrefix(id.Gvk)
		if p < 5 {
			p0 = append(p0, versionLessId)
		} else if p == 9 {
			p2 = append(p2, versionLessId)
		} else {
			p1 = append(p1, versionLessId)
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
