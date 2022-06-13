package kustomize

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestLastAppliedConfig(t *testing.T) {
	srcJSON := "{\"apiVersion\": \"v1\", \"kind\": \"Namespace\", \"metadata\": {\"name\": \"test-unit\"}}"
	km := &kManifest{}
	err := km.load([]byte(srcJSON))
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	setLastAppliedConfig(km, true)

	annotations := km.resource.GetAnnotations()
	count := len(annotations)
	if count != 1 {
		t.Errorf("TestLastAppliedConfig: incorrect number of annotations, got: %d, want: %d.", count, 1)
	}

	lac := getLastAppliedConfig(km.resource, true)
	if lac != srcJSON {
		t.Errorf("TestLastAppliedConfig: incorrect annotation value, got: %s, want: %s.", srcJSON, lac)
	}
}

func randomDataHelper(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func TestLastAppliedConfigCompressed(t *testing.T) {
	filler := randomDataHelper(256 * (1 << 10))
	srcJSON := fmt.Sprintf("{\"apiVersion\": \"v1\", \"kind\": \"ConfigMap\", \"metadata\": {\"name\": \"test-unit\", \"namespace\": \"test-unit\"}, \"data\": {\"payload\": %q}}", filler)

	km := &kManifest{}
	err := km.load([]byte(srcJSON))
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	setLastAppliedConfig(km, true)

	annotations := km.resource.GetAnnotations()

	_, ok := annotations[gzipLastAppliedConfigAnnotation]
	assert.Equal(t, true, ok, "TestLastAppliedConfigCompressed: did not have gzipLastAppliedConfig annotation")

	count := len(annotations)
	if count != 1 {
		t.Errorf("TestLastAppliedConfigCompressed: incorrect number of annotations, got: %d, want: %d.", count, 1)
	}

	lac := getLastAppliedConfig(km.resource, true)
	if lac != srcJSON {
		t.Errorf("TestLastAppliedConfigCompressed: incorrect annotation value, got: %s, want: %s.", srcJSON, lac)
	}
}

func TestLastAppliedConfigCompressionDisabled(t *testing.T) {
	filler := randomDataHelper(256 * (1 << 10))
	srcJSON := fmt.Sprintf("{\"apiVersion\": \"v1\", \"kind\": \"ConfigMap\", \"metadata\": {\"name\": \"test-unit\", \"namespace\": \"test-unit\"}, \"data\": {\"payload\": %q}}", filler)

	km := &kManifest{}
	err := km.load([]byte(srcJSON))
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	setLastAppliedConfig(km, false)

	annotations := km.resource.GetAnnotations()

	_, ok := annotations[lastAppliedConfigAnnotation]
	assert.Equal(t, true, ok, "TestLastAppliedConfigCompressionDisabled: did not have lastAppliedConfig annotation")

	_, ok = annotations[gzipLastAppliedConfigAnnotation]
	assert.Equal(t, false, ok, "TestLastAppliedConfigCompressionDisabled: found gzipLastAppliedConfig annotation unexpectedly")

	count := len(annotations)
	if count != 1 {
		t.Errorf("TestLastAppliedConfigCompressionDisabled: incorrect number of annotations, got: %d, want: %d.", count, 1)
	}

	lac := getLastAppliedConfig(km.resource, false)
	if lac != srcJSON {
		t.Errorf("TestLastAppliedConfigCompressionDisabled: incorrect annotation value, got: %s, want: %s.", srcJSON, lac)
	}
}

func TestGetPatchStrategicMergePatch1(t *testing.T) {
	kmo := kManifest{}
	kmm := kManifest{}
	kmc := kManifest{}
	kmo.load([]byte(testGetPatchStrategicMergePatch1OriginalJSON))
	kmm.load([]byte(testGetPatchStrategicMergePatch1ModifiedJSON))
	kmc.load([]byte(testGetPatchStrategicMergePatch1CurrentJSON))

	pt, p, err := getPatch(kmm.gvk(), kmo.json, kmm.json, kmc.json)
	assert.Equal(t, nil, err, nil)
	assert.Equal(t, testGetPatchStrategicMergePatch1PatchJSON, string(p), nil)
	assert.Equal(t, types.StrategicMergePatchType, pt, nil)
}

var testGetPatchStrategicMergePatch1OriginalJSON string = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-update-args-command\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"command\":[\"sleep\",\"infinity\"],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}\n"},"creationTimestamp":null,"labels":{"app":"test"},"name":"test","namespace":"test-update-args-command"},"spec":{"replicas":1,"selector":{"matchLabels":{"app":"test"}},"strategy":{},"template":{"metadata":{"creationTimestamp":null,"labels":{"app":"test"}},"spec":{"containers":[{"command":["sleep","infinity"],"image":"nginx","name":"nginx","resources":{}}]}}},"status":{}}`
var testGetPatchStrategicMergePatch1ModifiedJSON string = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-update-args-command\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"args\":[\"-V\"],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}\n"},"creationTimestamp":null,"labels":{"app":"test"},"name":"test","namespace":"test-update-args-command"},"spec":{"replicas":1,"selector":{"matchLabels":{"app":"test"}},"strategy":{},"template":{"metadata":{"creationTimestamp":null,"labels":{"app":"test"}},"spec":{"containers":[{"args":["-V"],"image":"nginx","name":"nginx","resources":{}}]}}},"status":{}}`
var testGetPatchStrategicMergePatch1CurrentJSON string = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"annotations":{"deployment.kubernetes.io/revision":"5","kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-update-args-command\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"command\":[\"sleep\",\"infinity\"],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}\n"},"creationTimestamp":"2021-04-10T09:26:33Z","generation":5,"labels":{"app":"test"},"managedFields":[{"apiVersion":"apps/v1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:deployment.kubernetes.io/revision":{}}},"f:status":{"f:availableReplicas":{},"f:conditions":{".":{},"k:{\"type\":\"Available\"}":{".":{},"f:lastTransitionTime":{},"f:lastUpdateTime":{},"f:message":{},"f:reason":{},"f:status":{},"f:type":{}},"k:{\"type\":\"Progressing\"}":{".":{},"f:lastTransitionTime":{},"f:lastUpdateTime":{},"f:message":{},"f:reason":{},"f:status":{},"f:type":{}}},"f:observedGeneration":{},"f:readyReplicas":{},"f:replicas":{},"f:updatedReplicas":{}}},"manager":"kube-controller-manager","operation":"Update","time":"2021-04-10T12:33:12Z"},{"apiVersion":"apps/v1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:kubectl.kubernetes.io/last-applied-configuration":{}},"f:labels":{".":{},"f:app":{}}},"f:spec":{"f:progressDeadlineSeconds":{},"f:replicas":{},"f:revisionHistoryLimit":{},"f:selector":{"f:matchLabels":{".":{},"f:app":{}}},"f:strategy":{"f:rollingUpdate":{".":{},"f:maxSurge":{},"f:maxUnavailable":{}},"f:type":{}},"f:template":{"f:metadata":{"f:labels":{".":{},"f:app":{}}},"f:spec":{"f:containers":{"k:{\"name\":\"nginx\"}":{".":{},"f:command":{},"f:image":{},"f:imagePullPolicy":{},"f:name":{},"f:resources":{},"f:terminationMessagePath":{},"f:terminationMessagePolicy":{}}},"f:dnsPolicy":{},"f:restartPolicy":{},"f:schedulerName":{},"f:securityContext":{},"f:terminationGracePeriodSeconds":{}}}}},"manager":"terraform-provider-kustomization","operation":"Update","time":"2021-04-10T12:33:12Z"}],"name":"test","namespace":"test-update-args-command","resourceVersion":"40682","selfLink":"/apis/apps/v1/namespaces/test-update-args-command/deployments/test","uid":"1dc64da3-0fbb-4e3d-8176-dcc606d9c19d"},"spec":{"progressDeadlineSeconds":600,"replicas":1,"revisionHistoryLimit":10,"selector":{"matchLabels":{"app":"test"}},"strategy":{"rollingUpdate":{"maxSurge":"25%","maxUnavailable":"25%"},"type":"RollingUpdate"},"template":{"metadata":{"creationTimestamp":null,"labels":{"app":"test"}},"spec":{"containers":[{"command":["sleep","infinity"],"image":"nginx","imagePullPolicy":"Always","name":"nginx","resources":{},"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File"}],"dnsPolicy":"ClusterFirst","restartPolicy":"Always","schedulerName":"default-scheduler","securityContext":{},"terminationGracePeriodSeconds":30}}},"status":{"availableReplicas":1,"conditions":[{"lastTransitionTime":"2021-04-10T09:26:36Z","lastUpdateTime":"2021-04-10T09:26:36Z","message":"Deployment has minimum availability.","reason":"MinimumReplicasAvailable","status":"True","type":"Available"},{"lastTransitionTime":"2021-04-10T09:52:24Z","lastUpdateTime":"2021-04-10T12:33:12Z","message":"ReplicaSet \"test-dfc596dc6\" has successfully progressed.","reason":"NewReplicaSetAvailable","status":"True","type":"Progressing"}],"observedGeneration":5,"readyReplicas":1,"replicas":1,"updatedReplicas":1}}`
var testGetPatchStrategicMergePatch1PatchJSON string = `{"metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-update-args-command\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"args\":[\"-V\"],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}\n"},"creationTimestamp":null},"spec":{"template":{"spec":{"$setElementOrder/containers":[{"name":"nginx"}],"containers":[{"args":["-V"],"command":null,"name":"nginx"}]}}}}`

func TestGetPatchStrategicMergePatch2(t *testing.T) {
	kmo := kManifest{}
	kmm := kManifest{}
	kmc := kManifest{}
	kmo.load([]byte(testGetPatchStrategicMergePatch2OriginalJSON))
	kmm.load([]byte(testGetPatchStrategicMergePatch2ModifiedJSON))
	kmc.load([]byte(testGetPatchStrategicMergePatch2CurrentJSON))

	pt, p, err := getPatch(kmm.gvk(), kmo.json, kmm.json, kmc.json)
	assert.Equal(t, nil, err, nil)
	assert.Equal(t, testGetPatchStrategicMergePatch2PatchJSON, string(p), nil)
	assert.Equal(t, types.StrategicMergePatchType, pt, nil)
}

var testGetPatchStrategicMergePatch2OriginalJSON string = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-update-args-command\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"command\":[\"sleep\",\"infinity\"],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}\n"},"creationTimestamp":null,"labels":{"app":"test"},"name":"test","namespace":"test-update-args-command"},"spec":{"replicas":1,"selector":{"matchLabels":{"app":"test"}},"strategy":{},"template":{"metadata":{"creationTimestamp":null,"labels":{"app":"test"}},"spec":{"containers":[{"command":["sleep","infinity"],"image":"nginx","name":"nginx","resources":{}}]}}},"status":{}}`
var testGetPatchStrategicMergePatch2ModifiedJSON string = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-update-args-command\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{\"type\":\"Recreate\"},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"command\":[\"sleep\",\"infinity\"],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}\n"},"creationTimestamp":null,"labels":{"app":"test"},"name":"test","namespace":"test-update-args-command"},"spec":{"replicas":1,"selector":{"matchLabels":{"app":"test"}},"strategy":{"type":"Recreate"},"template":{"metadata":{"creationTimestamp":null,"labels":{"app":"test"}},"spec":{"containers":[{"command":["sleep","infinity"],"image":"nginx","name":"nginx","resources":{}}]}}},"status":{}}`
var testGetPatchStrategicMergePatch2CurrentJSON string = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"annotations":{"deployment.kubernetes.io/revision":"5","kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-update-args-command\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"command\":[\"sleep\",\"infinity\"],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}\n"},"creationTimestamp":"2021-04-10T09:26:33Z","generation":5,"labels":{"app":"test"},"managedFields":[{"apiVersion":"apps/v1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:deployment.kubernetes.io/revision":{}}},"f:status":{"f:availableReplicas":{},"f:conditions":{".":{},"k:{\"type\":\"Available\"}":{".":{},"f:lastTransitionTime":{},"f:lastUpdateTime":{},"f:message":{},"f:reason":{},"f:status":{},"f:type":{}},"k:{\"type\":\"Progressing\"}":{".":{},"f:lastTransitionTime":{},"f:lastUpdateTime":{},"f:message":{},"f:reason":{},"f:status":{},"f:type":{}}},"f:observedGeneration":{},"f:readyReplicas":{},"f:replicas":{},"f:updatedReplicas":{}}},"manager":"kube-controller-manager","operation":"Update","time":"2021-04-10T12:33:12Z"},{"apiVersion":"apps/v1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:kubectl.kubernetes.io/last-applied-configuration":{}},"f:labels":{".":{},"f:app":{}}},"f:spec":{"f:progressDeadlineSeconds":{},"f:replicas":{},"f:revisionHistoryLimit":{},"f:selector":{"f:matchLabels":{".":{},"f:app":{}}},"f:strategy":{"f:rollingUpdate":{".":{},"f:maxSurge":{},"f:maxUnavailable":{}},"f:type":{}},"f:template":{"f:metadata":{"f:labels":{".":{},"f:app":{}}},"f:spec":{"f:containers":{"k:{\"name\":\"nginx\"}":{".":{},"f:command":{},"f:image":{},"f:imagePullPolicy":{},"f:name":{},"f:resources":{},"f:terminationMessagePath":{},"f:terminationMessagePolicy":{}}},"f:dnsPolicy":{},"f:restartPolicy":{},"f:schedulerName":{},"f:securityContext":{},"f:terminationGracePeriodSeconds":{}}}}},"manager":"terraform-provider-kustomization","operation":"Update","time":"2021-04-10T12:33:12Z"}],"name":"test","namespace":"test-update-args-command","resourceVersion":"40682","selfLink":"/apis/apps/v1/namespaces/test-update-args-command/deployments/test","uid":"1dc64da3-0fbb-4e3d-8176-dcc606d9c19d"},"spec":{"progressDeadlineSeconds":600,"replicas":1,"revisionHistoryLimit":10,"selector":{"matchLabels":{"app":"test"}},"strategy":{"rollingUpdate":{"maxSurge":"25%","maxUnavailable":"25%"},"type":"RollingUpdate"},"template":{"metadata":{"creationTimestamp":null,"labels":{"app":"test"}},"spec":{"containers":[{"command":["sleep","infinity"],"image":"nginx","imagePullPolicy":"Always","name":"nginx","resources":{},"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File"}],"dnsPolicy":"ClusterFirst","restartPolicy":"Always","schedulerName":"default-scheduler","securityContext":{},"terminationGracePeriodSeconds":30}}},"status":{"availableReplicas":1,"conditions":[{"lastTransitionTime":"2021-04-10T09:26:36Z","lastUpdateTime":"2021-04-10T09:26:36Z","message":"Deployment has minimum availability.","reason":"MinimumReplicasAvailable","status":"True","type":"Available"},{"lastTransitionTime":"2021-04-10T09:52:24Z","lastUpdateTime":"2021-04-10T12:33:12Z","message":"ReplicaSet \"test-dfc596dc6\" has successfully progressed.","reason":"NewReplicaSetAvailable","status":"True","type":"Progressing"}],"observedGeneration":5,"readyReplicas":1,"replicas":1,"updatedReplicas":1}}`
var testGetPatchStrategicMergePatch2PatchJSON string = `{"metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"},\"name\":\"test\",\"namespace\":\"test-update-args-command\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"strategy\":{\"type\":\"Recreate\"},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"test\"}},\"spec\":{\"containers\":[{\"command\":[\"sleep\",\"infinity\"],\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}\n"},"creationTimestamp":null},"spec":{"strategy":{"$retainKeys":["type"],"type":"Recreate"}}}`

func TestGetPatchMergePatch(t *testing.T) {
	kmo := kManifest{}
	kmm := kManifest{}
	kmc := kManifest{}
	kmo.load([]byte(`{"apiVersion":"test.example.com/v1alpha1","kind":"Namespacedcrd","metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"test.example.com/v1alpha1\",\"kind\":\"Namespacedcrd\",\"metadata\":{\"name\":\"namespacedco\",\"namespace\":\"test-crd\"},\"spec\":{}}\n"},"name":"namespacedco","namespace":"test-crd"},"spec":{}}`))
	kmm.load([]byte(`{"apiVersion":"test.example.com/v1alpha1","kind":"Namespacedcrd","metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"test.example.com/v1alpha1\",\"kind\":\"Namespacedcrd\",\"metadata\":{\"name\":\"namespacedco\",\"namespace\":\"test-crd\"},\"spec\":{\"test-key\":\"test-value\"}}\n"},"name":"namespacedco","namespace":"test-crd"},"spec":{"test-key":"test-value"}}`))
	kmc.load([]byte(`{"apiVersion":"test.example.com/v1alpha1","kind":"Namespacedcrd","metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"test.example.com/v1alpha1\",\"kind\":\"Namespacedcrd\",\"metadata\":{\"name\":\"namespacedco\",\"namespace\":\"test-crd\"},\"spec\":{}}\n"},"creationTimestamp":"2021-04-10T16:17:56Z","generation":1,"managedFields":[{"apiVersion":"test.example.com/v1alpha1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:kubectl.kubernetes.io/last-applied-configuration":{}}},"f:spec":{}},"manager":"terraform-provider-kustomization","operation":"Update","time":"2021-04-10T16:17:56Z"}],"name":"namespacedco","namespace":"test-crd","resourceVersion":"697","selfLink":"/apis/test.example.com/v1alpha1/namespaces/test-crd/namespacedcrds/namespacedco","uid":"fff6097d-acd4-4ec9-86ac-044d3e03f685"},"spec":{}}`))

	pt, p, err := getPatch(kmm.gvk(), kmo.json, kmm.json, kmc.json)
	assert.Equal(t, nil, err, nil)
	assert.Equal(t, `{"metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"test.example.com/v1alpha1\",\"kind\":\"Namespacedcrd\",\"metadata\":{\"name\":\"namespacedco\",\"namespace\":\"test-crd\"},\"spec\":{\"test-key\":\"test-value\"}}\n"}},"spec":{"test-key":"test-value"}}`, string(p), nil)
	assert.Equal(t, types.MergePatchType, pt, nil)
}
