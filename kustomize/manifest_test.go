package kustomize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKManifestLoad(t *testing.T) {
	ns := []byte(`{
		"kind": "Namespace",
		"apiVersion": "v1",
		"metadata": {
			"name": "test",
			"creationTimestamp": null
		},
		"spec": {},
		"status": {}
	}`)

	km := kManifest{}
	err := km.load(ns)

	assert.Equal(t, nil, err)
}

func TestKManifestLoadErr(t *testing.T) {
	ns := []byte("")

	km := kManifest{}
	err := km.load(ns)

	assert.NotEqual(t, nil, err)
}
