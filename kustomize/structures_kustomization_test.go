package kustomize

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
)

func TestFlattenKustomizationIDs(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	opts := krusty.MakeDefaultOptions()
	k := krusty.MakeKustomizer(fSys, opts)

	rm, err := k.Run("test_kustomizations/basic/initial")
	assert.Equal(t, err, nil, nil)

	ids, nsNotSet, nsSet := flattenKustomizationIDs(rm)

	expMerged := append(nsNotSet, nsSet...)
	assert.ElementsMatch(t, expMerged, ids, nil)

	expIds := []string{"~G_v1_Namespace|~X|test-basic", "apps_v1_Deployment|test-basic|test", "networking.k8s.io_v1beta1_Ingress|test-basic|test", "~G_v1_Service|test-basic|test"}
	assert.ElementsMatch(t, expIds, ids, nil)

	expNsNotSet := []string{"~G_v1_Namespace|~X|test-basic"}
	assert.ElementsMatch(t, expNsNotSet, nsNotSet, nil)

	expNsSet := []string{"apps_v1_Deployment|test-basic|test", "networking.k8s.io_v1beta1_Ingress|test-basic|test", "~G_v1_Service|test-basic|test"}
	assert.ElementsMatch(t, expNsSet, nsSet, nil)
}
