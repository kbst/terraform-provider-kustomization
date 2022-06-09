package kustomize

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeterminePrefix(t *testing.T) {
	for _, id := range residListFirst {
		kr, _ := parseProviderId(id)
		p := determinePrefix(kr)
		e := uint32(1)
		assert.Equal(t, e, p, nil)
	}

	for _, id := range residListDefault {
		kr, _ := parseProviderId(id)
		p := determinePrefix(kr)
		e := uint32(5)
		assert.Equal(t, e, p, nil)
	}

	for _, id := range residListLast {
		kr, _ := parseProviderId(id)
		p := determinePrefix(kr)
		e := uint32(9)
		assert.Equal(t, e, p, nil)
	}
}

func TestPrefixHash(t *testing.T) {
	ti := uint32(math.MaxInt32 / 1000)
	i := prefixHash(uint32(1), ti)
	assert.Equal(t, 100021474, i, nil)

	i = prefixHash(uint32(5), ti)
	assert.Equal(t, 500021474, i, nil)

	i = prefixHash(uint32(9), ti)
	assert.Equal(t, 900021474, i, nil)
}

func TestPrefixHashMaxInt(t *testing.T) {
	ti := uint32(math.MaxInt32)
	i := prefixHash(uint32(1), ti)
	assert.Equal(t, 121474836, i, nil)

	i = prefixHash(uint32(5), ti)
	assert.Equal(t, 521474836, i, nil)

	i = prefixHash(uint32(9), ti)
	assert.Equal(t, 921474836, i, nil)
}

func TestIdSetHash(t *testing.T) {
	idList := []string{}
	idList = append(idList, residListFirst...)
	idList = append(idList, residListDefault...)
	idList = append(idList, residListLast...)
	setIDs := []int{}

	for _, s := range idList {
		f := idSetHash(s)

		assert.NotContains(t, setIDs, f, nil)
		setIDs = append(setIDs, f)
	}
}

var residListFirst = []string{
	"_/Namespace/_/test",
	"apiextensions.k8s.io/CustomResourceDefinition/_/test",
	"apiextensions.k8s.io/CustomResourceDefinition/test-ns/test",
}

var residListDefault = []string{
	"autoscaling/HorizontalPodAutoscaler/_/test",
	"autoscaling/HorizontalPodAutoscaler/test-ns/test",
	"coordination.k8s.io/Lease/_/test",
	"coordination.k8s.io/Lease/test-ns/test",
	"discovery.k8s.io/EndpointSlice/_/test",
	"discovery.k8s.io/EndpointSlice/test-ns/test",
	"extensions/Ingress/_/test",
	"extensions/Ingress/test-ns/test",
	"rbac.authorization.k8s.io/ClusterRoleBinding/_/test",
	"rbac.authorization.k8s.io/ClusterRoleBinding/test-ns/test",
	"rbac.authorization.k8s.io/ClusterRole/_/test",
	"rbac.authorization.k8s.io/ClusterRole/test-ns/test",
	"rbac.authorization.k8s.io/RoleBinding/_/test",
	"rbac.authorization.k8s.io/RoleBinding/test-ns/test",
	"rbac.authorization.k8s.io/Role/_/test",
	"rbac.authorization.k8s.io/Role/test-ns/test",
	"apiregistration.k8s.io/APIService/_/test",
	"apiregistration.k8s.io/APIService/test-ns/test",
	"batch/CronJob/_/test",
	"batch/CronJob/test-ns/test",
	"batch/Job/_/test",
	"batch/Job/test-ns/test",
	"authorization.k8s.io/LocalSubjectAccessReview/_/test",
	"authorization.k8s.io/LocalSubjectAccessReview/test-ns/test",
	"authorization.k8s.io/SelfSubjectAccessReview/_/test",
	"authorization.k8s.io/SelfSubjectAccessReview/test-ns/test",
	"authorization.k8s.io/SelfSubjectRulesReview/_/test",
	"authorization.k8s.io/SelfSubjectRulesReview/test-ns/test",
	"authorization.k8s.io/SubjectAccessReview/_/test",
	"authorization.k8s.io/SubjectAccessReview/test-ns/test",
	"certificates.k8s.io/CertificateSigningRequest/_/test",
	"certificates.k8s.io/CertificateSigningRequest/test-ns/test",
	"events.k8s.io/Event/_/test",
	"events.k8s.io/Event/test-ns/test",
	"node.k8s.io/RuntimeClass/_/test",
	"node.k8s.io/RuntimeClass/test-ns/test",
	"policy/PodDisruptionBudget/_/test",
	"policy/PodDisruptionBudget/test-ns/test",
	"policy/PodSecurityPolicy/_/test",
	"policy/PodSecurityPolicy/test-ns/test",
	"apps/ControllerRevision/_/test",
	"apps/ControllerRevision/test-ns/test",
	"apps/DaemonSet/_/test",
	"apps/DaemonSet/test-ns/test",
	"apps/Deployment/_/test",
	"apps/Deployment/test-ns/test",
	"apps/ReplicaSet/_/test",
	"apps/ReplicaSet/test-ns/test",
	"apps/StatefulSet/_/test",
	"apps/StatefulSet/test-ns/test",
	"authentication.k8s.io/TokenReview/_/test",
	"authentication.k8s.io/TokenReview/test-ns/test",
	"networking.k8s.io/IngressClass/_/test",
	"networking.k8s.io/IngressClass/test-ns/test",
	"networking.k8s.io/Ingress/_/test",
	"networking.k8s.io/Ingress/test-ns/test",
	"networking.k8s.io/NetworkPolicy/_/test",
	"networking.k8s.io/NetworkPolicy/test-ns/test",
	"scheduling.k8s.io/PriorityClass/_/test",
	"scheduling.k8s.io/PriorityClass/test-ns/test",
	"storage.k8s.io/CSIDriver/_/test",
	"storage.k8s.io/CSIDriver/test-ns/test",
	"storage.k8s.io/CSINode/_/test",
	"storage.k8s.io/CSINode/test-ns/test",
	"storage.k8s.io/StorageClass/_/test",
	"storage.k8s.io/StorageClass/test-ns/test",
	"storage.k8s.io/VolumeAttachment/_/test",
	"storage.k8s.io/VolumeAttachment/test-ns/test",
	"_/Binding/_/test",
	"_/Binding/test-ns/test",
	"_/ComponentStatus/_/test",
	"_/ComponentStatus/test-ns/test",
	"_/ConfigMap/_/test",
	"_/ConfigMap/test-ns/test",
	"_/Endpoints/_/test",
	"_/Endpoints/test-ns/test",
	"_/Event/_/test",
	"_/Event/test-ns/test",
	"_/LimitRange/_/test",
	"_/LimitRange/test-ns/test",
	"_/Node/_/test",
	"_/PersistentVolumeClaim/_/test",
	"_/PersistentVolumeClaim/test-ns/test",
	"_/PersistentVolume/_/test",
	"_/PersistentVolume/test-ns/test",
	"_/Pod/_/test",
	"_/Pod/test-ns/test",
	"_/PodTemplate/_/test",
	"_/PodTemplate/test-ns/test",
	"_/ReplicationController/_/test",
	"_/ReplicationController/test-ns/test",
	"_/ResourceQuota/_/test",
	"_/ResourceQuota/test-ns/test",
	"_/Secret/_/test",
	"_/Secret/test-ns/test",
	"_/ServiceAccount/_/test",
	"_/ServiceAccount/test-ns/test",
	"_/Service/_/test",
	"_/Service/test-ns/test",
}

var residListLast = []string{
	"admissionregistration.k8s.io/MutatingWebhookConfiguration/_/test",
	"admissionregistration.k8s.io/MutatingWebhookConfiguration/test-ns/test",
	"admissionregistration.k8s.io/ValidatingWebhookConfiguration/_/test",
	"admissionregistration.k8s.io/ValidatingWebhookConfiguration/test-ns/test",
}
