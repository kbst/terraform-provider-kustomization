package kustomize

import (
	"sigs.k8s.io/kustomize/api/resmap"
)

func flattenKustomizationIDs(rm resmap.ResMap) (ids []string, idsPrio [][]string, err error) {
	p0 := []string{}
	p1 := []string{}
	p2 := []string{}
	for _, id := range rm.AllIds() {
		kr := &KubernetesResource{
			Group:     id.Group,
			Kind:      id.Kind,
			Namespace: id.Namespace,
			Name:      id.Name,
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
		kr := &KubernetesResource{
			Group:     r.CurId().Group,
			Kind:      r.CurId().Kind,
			Namespace: r.GetNamespace(),
			Name:      r.GetName(),
		}

		json, err := r.MarshalJSON()
		if err != nil {
			return nil, err
		}
		res[kr.toString()] = string(json)
	}
	return res, nil
}
