package kustomize

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestGetPatch(t *testing.T) {
	srcJSON := "{\"apiVersion\": \"v1\", \"kind\": \"Namespace\", \"metadata\": {\"name\": \"test-unit\"}}"

	o, _ := parseJSON(srcJSON)
	m, _ := parseJSON(srcJSON)
	c, _ := parseJSON(srcJSON)

	original, _ := o.MarshalJSON()
	modified, _ := m.MarshalJSON()
	current, _ := c.MarshalJSON()

	_, err := getPatch(original, modified, current)
	if err != nil {
		t.Errorf("TestGetPatch: %s", err)
	}
}

func TestSimplifyJSON(t *testing.T) {
	full := `{
  "apiVersion": "v1",
  "kind": "Service",
  "metadata": {
    "creationTimestamp": "2020-02-24T06:37:48Z",
    "labels": {
      "app": "foo"
    },
    "name": "foo",
    "namespace": "test",
    "resourceVersion": "70032962",
    "selfLink": "/api/v1/namespaces/test/services/foo",
    "uid": "2c0baf14-56d0-11ea-a787-42010a0e000b"
  },
  "spec": {
    "clusterIP": "10.12.38.10",
    "externalTrafficPolicy": "Cluster",
    "ports": [
      {
        "name": "http",
        "nodePort": 31033,
        "port": 8080,
        "protocol": "TCP",
        "targetPort": "http"
      }
    ],
    "selector": {
      "app": "foo"
    },
    "sessionAffinity": "None",
    "type": "NodePort"
  },
  "status": {
    "loadBalancer": {}
  }
}`

	targeted := `{
  "apiVersion": "v1",
  "kind": "Service",
  "metadata": {
    "name": "foo"
  },
  "spec": {
    "externalTrafficPolicy": "Cluster",
    "ports": [
      {
        "name": "http",
        "port": 7070,
        "protocol": "TCP",
        "targetPort": "http"
      },
      {
        "name": "monitoring",
        "port": 6666,
        "protocol": "TCP",
        "targetPort": "monitoring"
      }
    ],
    "selector": {
      "app": "bar"
    },
    "sessionAffinity": "None",
    "type": "NodePort"
  }
}`

	expected := `{
  "apiVersion": "v1",
  "kind": "Service",
  "metadata": {
    "name": "foo"
  },
  "spec": {
    "externalTrafficPolicy": "Cluster",
    "ports": [
      {
        "name": "http",
        "nodePort": 31033,
        "port": 8080,
        "protocol": "TCP",
        "targetPort": "http"
      }
    ],
    "selector": {
      "app": "foo"
    },
    "sessionAffinity": "None",
    "type": "NodePort"
  }
}`

	result, err := simplifyJSON([]byte(full), []byte(targeted))
	if err != nil {
		t.Errorf("TestSimplify: %s", err)
	}

	var resultPretty bytes.Buffer
	_ = json.Indent(&resultPretty, result, "", "  ")

	if expected != string(resultPretty.Bytes()) {
		t.Errorf("expected: %s\n got: %s", expected, resultPretty.Bytes())
	}
}
