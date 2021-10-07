package kustomize

import (
	"fmt"
	"regexp"

	"sigs.k8s.io/kustomize/kyaml/resid"
)

const RE_NOT_PIPE = "[^|]*"
const RE_NOT_UNDERSCORE = "[^_]*"
const RE_NOT_DQUOTE = "[^\"]*"
const RE_NOT_SLASH = "[^/]*"

const RE_KUSTOMIZATION_GROUP = "(?P<group>" + RE_NOT_UNDERSCORE + ")"
const RE_KUSTOMIZATION_API_VERSION = "(?P<version>" + RE_NOT_UNDERSCORE + ")"
const RE_KUSTOMIZATION_KIND = "(?P<kind>" + RE_NOT_PIPE + ")"
const RE_KUSTOMIZATION_NAMESPACE = "(?P<namespace>" + RE_NOT_PIPE + ")"
const RE_KUSTOMIZATION_NAME = "(?P<name>" + RE_NOT_DQUOTE + ")"

const RE_PROVIDER_GROUP = "(?P<group>" + RE_NOT_SLASH + ")"
const RE_PROVIDER_KIND = "(?P<kind>" + RE_NOT_SLASH + ")"
const RE_PROVIDER_NAMESPACE = "(?P<namespace>" + RE_NOT_SLASH + ")"
const RE_PROVIDER_NAME = "(?P<name>" + RE_NOT_SLASH + ")"

const RE_KUSTOMIZATION_ID = RE_KUSTOMIZATION_GROUP + "_" + RE_KUSTOMIZATION_API_VERSION + "_" + RE_KUSTOMIZATION_KIND + "\\|" +
	RE_KUSTOMIZATION_NAMESPACE + "\\|" + RE_KUSTOMIZATION_NAME
const RE_PROVIDER_ID = RE_PROVIDER_GROUP + "/" + RE_PROVIDER_KIND + "/" + RE_PROVIDER_NAMESPACE + "/" + RE_PROVIDER_NAME

var reKustomizationId = regexp.MustCompile(RE_KUSTOMIZATION_ID)
var reProviderId = regexp.MustCompile(RE_PROVIDER_ID)

type KubernetesResource struct {
	Group     string
	Kind      string
	Namespace string
	Name      string
}

func mustParseKustomizationId(id string) *KubernetesResource {
	rid := resid.FromString(id)

	return &KubernetesResource{
		Group:     rid.Gvk.Group,
		Kind:      rid.Gvk.Kind,
		Namespace: rid.Namespace,
		Name:      rid.Name,
	}
}

func parseKustomizationId(id string) (*KubernetesResource, error) {
	if !reKustomizationId.Match([]byte(id)) {
		return nil, fmt.Errorf("id %s does not match format group_version_Kind|namespace|name", id)
	}
	return mustParseKustomizationId(id), nil
}

func parseProviderId(id string) (*KubernetesResource, error) {
	match := reProviderId.FindStringSubmatch(id)
	if match == nil {
		return nil, fmt.Errorf("id %s does not match format group/Kind/namespace/name", id)
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

func convertKustomizeToTerraform(id string, legacy bool) (string, error) {
	if legacy {
		return id, nil
	}
	k, err := parseKustomizationId(id)
	if err != nil {
		return "", err
	} else {
		return k.toString(), nil
	}
}

func parseEitherIdFormat(id string) (*KubernetesResource, error) {
	k, err := parseProviderId(id)
	if err != nil {
		k, err = parseKustomizationId(id)
		if err != nil {
			return nil, fmt.Errorf("invalid ID: %q, valid IDs look like: \"_/Namespace/_/example\" or \"~G_v1_Namespace|~X|example\"", id)
		}
	}
	return k, nil
}

func mustParseEitherIdFormat(id string) *KubernetesResource {
	k, err := parseProviderId(id)
	if err != nil {
		return mustParseKustomizationId(id)
	}
	return k
}
