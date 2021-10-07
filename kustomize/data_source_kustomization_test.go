package kustomize

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeterminePrefix(t *testing.T) {
	for _, id := range residListFirst {
		kr := mustParseEitherIdFormat(id)
		p := determinePrefix(kr)
		e := uint32(1)
		assert.Equal(t, e, p, nil)
	}

	for _, id := range residListDefault {
		kr := mustParseEitherIdFormat(id)
		p := determinePrefix(kr)
		e := uint32(5)
		assert.Equal(t, e, p, nil)
	}

	for _, id := range residListLast {
		kr := mustParseEitherIdFormat(id)
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
	"~G_v1_Namespace|~X|test",
	"apiextensions.k8s.io_v1_CustomResourceDefinition|~X|test",
	"apiextensions.k8s.io_v1_CustomResourceDefinition|test-ns|test",
	"apiextensions.k8s.io_v1beta1_CustomResourceDefinition|~X|test",
	"apiextensions.k8s.io_v1beta1_CustomResourceDefinition|test-ns|test",
}

var residListDefault = []string{
	"autoscaling_v1_HorizontalPodAutoscaler|~X|test",
	"autoscaling_v1_HorizontalPodAutoscaler|test-ns|test",
	"autoscaling_v2beta1_HorizontalPodAutoscaler|~X|test",
	"autoscaling_v2beta1_HorizontalPodAutoscaler|test-ns|test",
	"autoscaling_v2beta2_HorizontalPodAutoscaler|~X|test",
	"autoscaling_v2beta2_HorizontalPodAutoscaler|test-ns|test",
	"coordination.k8s.io_v1_Lease|~X|test",
	"coordination.k8s.io_v1_Lease|test-ns|test",
	"coordination.k8s.io_v1beta1_Lease|~X|test",
	"coordination.k8s.io_v1beta1_Lease|test-ns|test",
	"discovery.k8s.io_v1beta1_EndpointSlice|~X|test",
	"discovery.k8s.io_v1beta1_EndpointSlice|test-ns|test",
	"extensions_v1beta1_Ingress|~X|test",
	"extensions_v1beta1_Ingress|test-ns|test",
	"rbac.authorization.k8s.io_v1_ClusterRoleBinding|~X|test",
	"rbac.authorization.k8s.io_v1_ClusterRoleBinding|test-ns|test",
	"rbac.authorization.k8s.io_v1beta1_ClusterRoleBinding|~X|test",
	"rbac.authorization.k8s.io_v1beta1_ClusterRoleBinding|test-ns|test",
	"rbac.authorization.k8s.io_v1_ClusterRole|~X|test",
	"rbac.authorization.k8s.io_v1_ClusterRole|test-ns|test",
	"rbac.authorization.k8s.io_v1beta1_ClusterRole|~X|test",
	"rbac.authorization.k8s.io_v1beta1_ClusterRole|test-ns|test",
	"rbac.authorization.k8s.io_v1_RoleBinding|~X|test",
	"rbac.authorization.k8s.io_v1_RoleBinding|test-ns|test",
	"rbac.authorization.k8s.io_v1beta1_RoleBinding|~X|test",
	"rbac.authorization.k8s.io_v1beta1_RoleBinding|test-ns|test",
	"rbac.authorization.k8s.io_v1_Role|~X|test",
	"rbac.authorization.k8s.io_v1_Role|test-ns|test",
	"rbac.authorization.k8s.io_v1beta1_Role|~X|test",
	"rbac.authorization.k8s.io_v1beta1_Role|test-ns|test",
	"apiregistration.k8s.io_v1_APIService|~X|test",
	"apiregistration.k8s.io_v1_APIService|test-ns|test",
	"apiregistration.k8s.io_v1beta1_APIService|~X|test",
	"apiregistration.k8s.io_v1beta1_APIService|test-ns|test",
	"batch_v1_CronJob|~X|test",
	"batch_v1_CronJob|test-ns|test",
	"batch_v1beta1_CronJob|~X|test",
	"batch_v1beta1_CronJob|test-ns|test",
	"batch_v1_Job|~X|test",
	"batch_v1_Job|test-ns|test",
	"batch_v1beta1_Job|~X|test",
	"batch_v1beta1_Job|test-ns|test",
	"authorization.k8s.io_v1_LocalSubjectAccessReview|~X|test",
	"authorization.k8s.io_v1_LocalSubjectAccessReview|test-ns|test",
	"authorization.k8s.io_v1beta1_LocalSubjectAccessReview|~X|test",
	"authorization.k8s.io_v1beta1_LocalSubjectAccessReview|test-ns|test",
	"authorization.k8s.io_v1_SelfSubjectAccessReview|~X|test",
	"authorization.k8s.io_v1_SelfSubjectAccessReview|test-ns|test",
	"authorization.k8s.io_v1beta1_SelfSubjectAccessReview|~X|test",
	"authorization.k8s.io_v1beta1_SelfSubjectAccessReview|test-ns|test",
	"authorization.k8s.io_v1_SelfSubjectRulesReview|~X|test",
	"authorization.k8s.io_v1_SelfSubjectRulesReview|test-ns|test",
	"authorization.k8s.io_v1beta1_SelfSubjectRulesReview|~X|test",
	"authorization.k8s.io_v1beta1_SelfSubjectRulesReview|test-ns|test",
	"authorization.k8s.io_v1_SubjectAccessReview|~X|test",
	"authorization.k8s.io_v1_SubjectAccessReview|test-ns|test",
	"authorization.k8s.io_v1beta1_SubjectAccessReview|~X|test",
	"authorization.k8s.io_v1beta1_SubjectAccessReview|test-ns|test",
	"certificates.k8s.io_v1beta1_CertificateSigningRequest|~X|test",
	"certificates.k8s.io_v1beta1_CertificateSigningRequest|test-ns|test",
	"events.k8s.io_v1beta1_Event|~X|test",
	"events.k8s.io_v1beta1_Event|test-ns|test",
	"node.k8s.io_v1beta1_RuntimeClass|~X|test",
	"node.k8s.io_v1beta1_RuntimeClass|test-ns|test",
	"policy_v1beta1_PodDisruptionBudget|~X|test",
	"policy_v1beta1_PodDisruptionBudget|test-ns|test",
	"policy_v1beta1_PodSecurityPolicy|~X|test",
	"policy_v1beta1_PodSecurityPolicy|test-ns|test",
	"apps_v1_ControllerRevision|~X|test",
	"apps_v1_ControllerRevision|test-ns|test",
	"apps_v1_DaemonSet|~X|test",
	"apps_v1_DaemonSet|test-ns|test",
	"apps_v1_Deployment|~X|test",
	"apps_v1_Deployment|test-ns|test",
	"apps_v1_ReplicaSet|~X|test",
	"apps_v1_ReplicaSet|test-ns|test",
	"apps_v1_StatefulSet|~X|test",
	"apps_v1_StatefulSet|test-ns|test",
	"authentication.k8s.io_v1_TokenReview|~X|test",
	"authentication.k8s.io_v1_TokenReview|test-ns|test",
	"authentication.k8s.io_v1beta1_TokenReview|~X|test",
	"authentication.k8s.io_v1beta1_TokenReview|test-ns|test",
	"networking.k8s.io_v1_IngressClass|~X|test",
	"networking.k8s.io_v1_IngressClass|test-ns|test",
	"networking.k8s.io_v1beta1_IngressClass|~X|test",
	"networking.k8s.io_v1beta1_IngressClass|test-ns|test",
	"networking.k8s.io_v1_Ingress|~X|test",
	"networking.k8s.io_v1_Ingress|test-ns|test",
	"networking.k8s.io_v1beta1_Ingress|~X|test",
	"networking.k8s.io_v1beta1_Ingress|test-ns|test",
	"networking.k8s.io_v1_NetworkPolicy|~X|test",
	"networking.k8s.io_v1_NetworkPolicy|test-ns|test",
	"networking.k8s.io_v1beta1_NetworkPolicy|~X|test",
	"networking.k8s.io_v1beta1_NetworkPolicy|test-ns|test",
	"scheduling.k8s.io_v1_PriorityClass|~X|test",
	"scheduling.k8s.io_v1_PriorityClass|test-ns|test",
	"scheduling.k8s.io_v1beta1_PriorityClass|~X|test",
	"scheduling.k8s.io_v1beta1_PriorityClass|test-ns|test",
	"storage.k8s.io_v1_CSIDriver|~X|test",
	"storage.k8s.io_v1_CSIDriver|test-ns|test",
	"storage.k8s.io_v1beta1_CSIDriver|~X|test",
	"storage.k8s.io_v1beta1_CSIDriver|test-ns|test",
	"storage.k8s.io_v1_CSINode|~X|test",
	"storage.k8s.io_v1_CSINode|test-ns|test",
	"storage.k8s.io_v1beta1_CSINode|~X|test",
	"storage.k8s.io_v1beta1_CSINode|test-ns|test",
	"storage.k8s.io_v1_StorageClass|~X|test",
	"storage.k8s.io_v1_StorageClass|test-ns|test",
	"storage.k8s.io_v1beta1_StorageClass|~X|test",
	"storage.k8s.io_v1beta1_StorageClass|test-ns|test",
	"storage.k8s.io_v1_VolumeAttachment|~X|test",
	"storage.k8s.io_v1_VolumeAttachment|test-ns|test",
	"storage.k8s.io_v1beta1_VolumeAttachment|~X|test",
	"storage.k8s.io_v1beta1_VolumeAttachment|test-ns|test",
	"~G_v1_Binding|~X|test",
	"~G_v1_Binding|test-ns|test",
	"~G_v1_ComponentStatus|~X|test",
	"~G_v1_ComponentStatus|test-ns|test",
	"~G_v1_ConfigMap|~X|test",
	"~G_v1_ConfigMap|test-ns|test",
	"~G_v1_Endpoints|~X|test",
	"~G_v1_Endpoints|test-ns|test",
	"~G_v1_Event|~X|test",
	"~G_v1_Event|test-ns|test",
	"~G_v1_LimitRange|~X|test",
	"~G_v1_LimitRange|test-ns|test",
	"~G_v1_Node|~X|test",
	"~G_v1_PersistentVolumeClaim|~X|test",
	"~G_v1_PersistentVolumeClaim|test-ns|test",
	"~G_v1_PersistentVolume|~X|test",
	"~G_v1_PersistentVolume|test-ns|test",
	"~G_v1_Pod|~X|test",
	"~G_v1_Pod|test-ns|test",
	"~G_v1_PodTemplate|~X|test",
	"~G_v1_PodTemplate|test-ns|test",
	"~G_v1_ReplicationController|~X|test",
	"~G_v1_ReplicationController|test-ns|test",
	"~G_v1_ResourceQuota|~X|test",
	"~G_v1_ResourceQuota|test-ns|test",
	"~G_v1_Secret|~X|test",
	"~G_v1_Secret|test-ns|test",
	"~G_v1_ServiceAccount|~X|test",
	"~G_v1_ServiceAccount|test-ns|test",
	"~G_v1_Service|~X|test",
	"~G_v1_Service|test-ns|test",
}

var residListLast = []string{
	"admissionregistration.k8s.io_v1_MutatingWebhookConfiguration|~X|test",
	"admissionregistration.k8s.io_v1_MutatingWebhookConfiguration|test-ns|test",
	"admissionregistration.k8s.io_v1beta1_MutatingWebhookConfiguration|~X|test",
	"admissionregistration.k8s.io_v1beta1_MutatingWebhookConfiguration|test-ns|test",
	"admissionregistration.k8s.io_v1_ValidatingWebhookConfiguration|~X|test",
	"admissionregistration.k8s.io_v1_ValidatingWebhookConfiguration|test-ns|test",
	"admissionregistration.k8s.io_v1beta1_ValidatingWebhookConfiguration|~X|test",
	"admissionregistration.k8s.io_v1beta1_ValidatingWebhookConfiguration|test-ns|test",
}
