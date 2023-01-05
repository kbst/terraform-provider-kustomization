package kustomize

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	k8sappsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smeta "k8s.io/apimachinery/pkg/api/meta"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sunstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	k8sdynamic "k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

var waitRefreshFunctions = map[string]waitRefreshFunction{
	"apps/Deployment": waitDeploymentRefresh,
	"apps/Daemonset":  waitDaemonsetRefresh,
}

type kManifestId struct {
	group     string
	kind      string
	namespace string
	name      string
}

type waitRefreshFunction func(km *kManifest) (interface{}, string, error)

func mustParseProviderId(str string) *kManifestId {
	kr, err := parseProviderId(str)
	if err != nil {
		log.Fatal(err)
	}

	return kr
}

func parseProviderId(str string) (*kManifestId, error) {
	parts := strings.Split(str, "/")

	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid ID: %q, valid IDs look like: \"_/Namespace/_/example\"", str)
	}

	group := parts[0]
	kind := parts[1]
	namespace := parts[2]
	name := parts[3]

	return &kManifestId{
		group:     underscoreToEmpty(group),
		kind:      kind,
		namespace: underscoreToEmpty(namespace),
		name:      name,
	}, nil
}

func (k kManifestId) string() string {
	return fmt.Sprintf("%s/%s/%s/%s", emptyToUnderscore(k.group), k.kind, emptyToUnderscore(k.namespace), k.name)
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

type kManifest struct {
	resource *k8sunstructured.Unstructured
	mapper   *restmapper.DeferredDiscoveryRESTMapper
	client   k8sdynamic.Interface
	json     []byte
}

func newKManifest(mapper *restmapper.DeferredDiscoveryRESTMapper, client k8sdynamic.Interface) *kManifest {
	return &kManifest{
		mapper: mapper,
		client: client,
	}
}

func (km *kManifest) load(body []byte) error {
	obj, err := k8sruntime.Decode(k8sunstructured.UnstructuredJSONScheme, body)
	if err != nil {
		return logError(fmt.Errorf("json error: %s", err))
	}

	km.resource = obj.(*k8sunstructured.Unstructured)
	km.json = body

	return nil
}

func (km *kManifest) gvk() k8sschema.GroupVersionKind {
	return km.resource.GroupVersionKind()
}

func (km *kManifest) gvr() (k8sschema.GroupVersionResource, error) {
	m, err := km.mapping()
	return m.Resource, err
}

func (km *kManifest) mapping() (m *k8smeta.RESTMapping, err error) {
	return km.mapper.RESTMapping(km.gvk().GroupKind(), km.gvk().Version)
}

func (km *kManifest) mappings() (m []*k8smeta.RESTMapping, err error) {
	return km.mapper.RESTMappings(km.gvk().GroupKind())
}

func (km *kManifest) namespace() string {
	return km.id().namespace
}

func (km *kManifest) isNamespaced() (bool, error) {
	m, err := km.mapping()
	if err != nil {
		return false, km.fmtErr(fmt.Errorf("api error: %s", err))
	}

	if m.Scope.Name() == k8smeta.RESTScopeNameNamespace {
		return true, nil
	}

	return false, nil
}

func (km *kManifest) name() string {
	return km.id().name
}

func (km *kManifest) id() (id kManifestId) {
	id.group = km.gvk().Group
	id.kind = km.gvk().Kind
	id.namespace = km.resource.GetNamespace()
	id.name = km.resource.GetName()

	return id
}

func (km *kManifest) api() (api k8sdynamic.ResourceInterface, err error) {
	gvr, err := km.gvr()
	if err != nil {
		return api, km.fmtErr(fmt.Errorf("api error: %s", err))
	}

	api = km.client.Resource(gvr)

	isNamespaced, err := km.isNamespaced()
	if err != nil {
		return api, km.fmtErr(fmt.Errorf("api error: %s", err))
	}

	if isNamespaced {
		api = km.client.Resource(gvr).Namespace(km.namespace())
	}

	return api, nil
}

func (km *kManifest) apiGet(opts k8smetav1.GetOptions) (resp *k8sunstructured.Unstructured, err error) {
	api, err := km.api()
	if err != nil {
		return resp, km.fmtErr(fmt.Errorf("get failed: %s", err))
	}

	return api.Get(context.TODO(), km.name(), opts)
}

func (km *kManifest) apiCreate(opts k8smetav1.CreateOptions) (resp *k8sunstructured.Unstructured, err error) {
	api, err := km.api()
	if err != nil {
		return resp, km.fmtErr(fmt.Errorf("create failed: %s", err))
	}

	return api.Create(context.TODO(), km.resource, opts)
}

func (km *kManifest) apiDelete(opts k8smetav1.DeleteOptions) (err error) {
	api, err := km.api()
	if err != nil {
		return km.fmtErr(fmt.Errorf("delete failed: %s", err))
	}

	return api.Delete(context.TODO(), km.name(), opts)
}

func (km *kManifest) apiPreparePatch(kmo *kManifest, currAllowNotFound bool) (pt k8stypes.PatchType, p []byte, err error) {
	original := kmo.json
	modified := km.json

	resp, err := km.apiGet(k8smetav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) || (k8serrors.IsNotFound(err) && !currAllowNotFound) {
			return pt, p, km.fmtErr(fmt.Errorf("error preparing patch: %s", err))
		}
	}

	current, err := resp.MarshalJSON()
	if err != nil {
		return pt, p, km.fmtErr(fmt.Errorf("error preparing patch: %s", err))
	}

	pt, p, err = getPatch(km.gvk(), original, modified, current)
	if err != nil {
		return pt, p, km.fmtErr(fmt.Errorf("error preparing patch: %s", err))
	}

	return pt, p, nil
}

func (km *kManifest) apiPatch(pt k8stypes.PatchType, p []byte, opts k8smetav1.PatchOptions) (resp *k8sunstructured.Unstructured, err error) {
	api, err := km.api()
	if err != nil {
		return resp, km.fmtErr(fmt.Errorf("patch failed: %s", err))
	}

	return api.Patch(context.TODO(), km.name(), pt, p, opts)
}

func parseResourceData(km *kManifest, d string) (err error) {
	b := []byte(d)

	err = km.load(b)
	if err != nil {
		return logError(fmt.Errorf("error parsing resource data: %s", err))
	}

	return nil
}

func (km *kManifest) getNamespaceManifest() (kns *kManifest, namespaced bool) {
	if km.namespace() == "" {
		return kns, false
	}

	kns = newKManifest(km.mapper, km.client)

	kns.resource = kns.resource.NewEmptyInstance().(*k8sunstructured.Unstructured)

	gvk := k8sschema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Namespace"}

	kns.resource.SetGroupVersionKind(gvk)
	kns.resource.SetName(km.namespace())

	return kns, true
}

func (km *kManifest) waitKind(t time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"existing"},
		Pending: []string{"pending"},
		Timeout: t,
		Refresh: func() (interface{}, string, error) {
			mapping, err := km.mapping()
			if err != nil {
				if k8smeta.IsNoMatchError(err) {
					// if not found, reset mapper cache
					// before trying again (required for CRDs)
					km.mapper.Reset()
					return nil, "pending", nil
				}
				return nil, "", err
			}

			return mapping.Resource, "existing", nil
		},
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return km.fmtErr(fmt.Errorf("timed out waiting for: %q: %s", km.gvk().String(), err))
	}

	return nil
}

func (km *kManifest) waitNamespace(t time.Duration) error {
	kns, namespaced := km.getNamespaceManifest()
	if !namespaced {
		// if the resource is not namespaced
		// we don't have to wait for it
		return nil
	}

	// wait for the namespace to exist
	_, err := kns.mappings()
	if err != nil {
		return km.fmtErr(
			fmt.Errorf("api error %q: %s", kns.id().string(), err),
		)
	}

	stateConf := &resource.StateChangeConf{
		Target:  []string{"existing"},
		Pending: []string{"pending"},
		Timeout: t,
		Refresh: func() (interface{}, string, error) {
			resp, err := kns.apiGet(k8smetav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return nil, "pending", nil
				}
				return nil, "", err
			}

			return resp, "existing", nil
		},
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return km.fmtErr(fmt.Errorf("timed out waiting for: %q: %s", kns.id().string(), err))
	}

	return nil
}

func (km *kManifest) waitDeleted(t time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Target:  []string{},
		Pending: []string{"deleting"},
		Timeout: t,
		Refresh: func() (interface{}, string, error) {
			resp, err := km.apiGet(k8smetav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return nil, "", nil
				}
				return nil, "", err
			}

			return resp, "deleting", nil
		},
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return km.fmtErr(fmt.Errorf("timed out deleting: %s", err))
	}

	return nil
}

func daemonsetReady(u *k8sunstructured.Unstructured) (bool, error) {
	var daemonset k8sappsv1.DaemonSet
	if err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &daemonset); err != nil {
		return false, err
	}
	if daemonset.Generation == daemonset.Status.ObservedGeneration &&
		daemonset.Status.UpdatedNumberScheduled == daemonset.Status.DesiredNumberScheduled &&
		daemonset.Status.NumberReady == daemonset.Status.DesiredNumberScheduled &&
		daemonset.Status.NumberUnavailable == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func waitDaemonsetRefresh(km *kManifest) (interface{}, string, error) {
	resp, err := km.apiGet(k8smetav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, "missing", nil
		}
		return nil, "error", err
	}
	ready, err := daemonsetReady(resp)
	if err != nil {
		return nil, "error", err
	}
	if ready {
		return resp, "done", nil
	}
	return nil, "in progress", nil
}

func deploymentReady(u *k8sunstructured.Unstructured) (bool, error) {
	var deployment k8sappsv1.Deployment
	if err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &deployment); err != nil {
		return false, err
	}
	if deployment.Generation == deployment.Status.ObservedGeneration &&
		deployment.Status.AvailableReplicas == *deployment.Spec.Replicas &&
		deployment.Status.AvailableReplicas == deployment.Status.Replicas &&
		deployment.Status.UnavailableReplicas == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func waitDeploymentRefresh(km *kManifest) (interface{}, string, error) {
	resp, err := km.apiGet(k8smetav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, "missing", nil
		}
		return nil, "error", err
	}
	ready, err := deploymentReady(resp)
	if err != nil {
		return nil, "error", err
	}
	if ready {
		return resp, "done", nil
	}
	return nil, "in progress", nil
}

func (km *kManifest) waitCreatedOrUpdated(t time.Duration) error {
	gvk := km.gvk()
	if refresh, ok := waitRefreshFunctions[fmt.Sprintf("%s/%s", gvk.Group, gvk.Kind)]; ok {
		delay := 10 * time.Second
		stateConf := &resource.StateChangeConf{
			Target:         []string{"done"},
			Pending:        []string{"in progress"},
			Timeout:        t,
			Delay:          delay,
			NotFoundChecks: 2*int(t/delay) + 1,
			Refresh: func() (interface{}, string, error) {
				return refresh(km)
			},
		}

		_, err := stateConf.WaitForState()
		if err != nil {
			return km.fmtErr(fmt.Errorf("timed out creating/updating %s %s/%s: %s", gvk.Kind, km.namespace(), km.name(), err))
		}
	}
	return nil
}

func (km *kManifest) fmtErr(err error) error {
	return fmt.Errorf(
		"%q: %s",
		km.id().string(),
		err)
}
