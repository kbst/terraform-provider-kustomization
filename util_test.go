package main

import (
	"testing"
)

func TestLastAppliedConfig(t *testing.T) {
	srcJSON := "{\"apiVersion\": \"v1\", \"kind\": \"Namespace\", \"metadata\": {\"name\": \"test-unit\"}}"
	u, err := parseJSON(srcJSON)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	setLastAppliedConfig(u, srcJSON)

	annotations := u.GetAnnotations()
	count := len(annotations)
	if count != 1 {
		t.Errorf("TestLastAppliedConfig: incorect number of annotations, got: %d, want: %d.", count, 1)
	}

	lac := getLastAppliedConfig(u)
	if lac != srcJSON {
		t.Errorf("TestLastAppliedConfig: incorect annotation value, got: %s, want: %s.", srcJSON, lac)
	}
}

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
