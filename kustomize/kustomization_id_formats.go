package kustomize

import (
	"fmt"
	"log"
	"strings"
)

type K8sResId struct {
	Group     string
	Kind      string
	Namespace string
	Name      string
}

func mustParseProviderId(str string) *K8sResId {
	kr, err := parseProviderId(str)
	if err != nil {
		log.Fatal(err)
	}

	return kr
}

func parseProviderId(str string) (*K8sResId, error) {
	parts := strings.Split(str, "/")

	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid ID: %q, valid IDs look like: \"_/Namespace/_/example\"", str)
	}

	group := parts[0]
	kind := parts[1]
	namespace := parts[2]
	name := parts[3]

	return &K8sResId{
		Group:     underscoreToEmpty(group),
		Kind:      kind,
		Namespace: underscoreToEmpty(namespace),
		Name:      name,
	}, nil
}

func (k K8sResId) toString() string {
	return fmt.Sprintf("%s/%s/%s/%s", emptyToUnderscore(k.Group), k.Kind, emptyToUnderscore(k.Namespace), k.Name)
}

func underscoreToEmpty(value string) string {
	if value == "_" {
		return ""
	}
	return value
}

func emptyToUnderscore(value string) string {
	if value == "" {
		return "_"
	}
	return value
}
