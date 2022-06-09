package kustomize

import (
	"fmt"
	"regexp"
)

const RE_NOT_SLASH = "[^/]*"

const RE_PROVIDER_GROUP = "(?P<group>" + RE_NOT_SLASH + ")"
const RE_PROVIDER_KIND = "(?P<kind>" + RE_NOT_SLASH + ")"
const RE_PROVIDER_NAMESPACE = "(?P<namespace>" + RE_NOT_SLASH + ")"
const RE_PROVIDER_NAME = "(?P<name>" + RE_NOT_SLASH + ")"

const RE_PROVIDER_ID = RE_PROVIDER_GROUP + "/" + RE_PROVIDER_KIND + "/" + RE_PROVIDER_NAMESPACE + "/" + RE_PROVIDER_NAME

var reProviderId = regexp.MustCompile(RE_PROVIDER_ID)

type KubernetesResource struct {
	Group     string
	Kind      string
	Namespace string
	Name      string
}

func parseProviderId(id string) (*KubernetesResource, error) {
	match := reProviderId.FindStringSubmatch(id)
	if match == nil {
		return nil, fmt.Errorf("invalid ID: %q, valid IDs look like: \"_/Namespace/_/example\"", id)
	}
	result := make(map[string]string)
	for i, name := range reProviderId.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	group := result["group"]
	if group == "_" {
		group = ""
	}
	namespace := result["namespace"]
	if namespace == "_" {
		namespace = ""
	}
	return &KubernetesResource{
		Group:     group,
		Kind:      result["kind"],
		Namespace: namespace,
		Name:      result["name"],
	}, nil
}

func emptyToUnderscore(value string) string {
	if value == "" {
		return "_"
	}
	return value
}

func (k KubernetesResource) toString() string {
	return fmt.Sprintf("%s/%s/%s/%s", emptyToUnderscore(k.Group), k.Kind, emptyToUnderscore(k.Namespace), k.Name)
}
