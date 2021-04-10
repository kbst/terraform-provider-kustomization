package kustomize

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"sigs.k8s.io/kustomize/api/resid"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smeta "k8s.io/apimachinery/pkg/api/meta"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func kustomizationResource() *schema.Resource {
	return &schema.Resource{
		Create:        kustomizationResourceCreate,
		Read:          kustomizationResourceRead,
		Exists:        kustomizationResourceExists,
		Update:        kustomizationResourceUpdate,
		Delete:        kustomizationResourceDelete,
		CustomizeDiff: kustomizationResourceDiff,

		Importer: &schema.ResourceImporter{
			State: kustomizationResourceImport,
		},

		Schema: map[string]*schema.Schema{
			"manifest": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func kustomizationResourceCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	srcJSON := d.Get("manifest").(string)
	u, err := parseJSON(srcJSON)
	if err != nil {
		return logError(fmt.Errorf("JSON parse error: %s", err))
	}

	stateConf := &resource.StateChangeConf{
		Target:  []string{"existing"},
		Pending: []string{"pending"},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: func() (interface{}, string, error) {
			// CRDs: wait for GroupVersionKind to exist
			gvr, err := cgvk.getGVR(u.GroupVersionKind(), true)
			if err != nil {
				if k8smeta.IsNoMatchError(err) {
					return nil, "pending", nil
				}
				return nil, "", err
			}

			return gvr, "existing", nil
		},
	}
	gvrResp, err := stateConf.WaitForState()
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("timed out waiting for apiVersion: %q, kind: %q to exist: %s", u.GroupVersionKind().GroupVersion(), u.GroupVersionKind().Kind, err),
		)
	}

	gvr := gvrResp.(k8sschema.GroupVersionResource)

	namespace := u.GetNamespace()

	setLastAppliedConfig(u, srcJSON)

	if namespace != "" {
		// wait for the namespace to exist
		nsGvk := k8sschema.GroupVersionKind{
			Group:   "",
			Version: "",
			Kind:    "Namespace"}
		nsGvr, err := cgvk.getGVR(nsGvk, false)
		if err != nil {
			return logErrorForResource(
				u,
				fmt.Errorf("api server has no apiVersion: %q, kind: %q: %s", nsGvk.GroupVersion(), nsGvk.Kind, err),
			)
		}

		stateConf := &resource.StateChangeConf{
			Target:  []string{"existing"},
			Pending: []string{"pending"},
			Timeout: d.Timeout(schema.TimeoutCreate),
			Refresh: func() (interface{}, string, error) {
				resp, err := client.
					Resource(nsGvr).
					Get(context.TODO(), namespace, k8smetav1.GetOptions{})
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
			return logErrorForResource(
				u,
				fmt.Errorf("timed out waiting for apiVersion: %q, kind: %q, name: %q, to exist: %s", nsGvk.GroupVersion(), nsGvk.Kind, namespace, err),
			)
		}
	}

	resp, err := client.
		Resource(gvr).
		Namespace(namespace).
		Create(context.TODO(), u, k8smetav1.CreateOptions{})
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("create failed: %s", err),
		)
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(resp))

	return kustomizationResourceRead(d, m)
}

func kustomizationResourceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	srcJSON := d.Get("manifest").(string)
	u, err := parseJSON(srcJSON)
	if err != nil {
		return logError(fmt.Errorf("JSON parse error: %s", err))
	}

	gvr, err := cgvk.getGVR(u.GroupVersionKind(), false)
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("failed to query GVR: %s", err),
		)
	}

	resp, err := client.
		Resource(gvr).
		Namespace(u.GetNamespace()).
		Get(context.TODO(), u.GetName(), k8smetav1.GetOptions{})
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("get failed: %s", err),
		)
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(resp))

	return nil
}

func kustomizationResourceDiff(d *schema.ResourceDiff, m interface{}) error {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	originalJSON, modifiedJSON := d.GetChange("manifest")

	if !d.HasChange("manifest") {
		return nil
	}

	srcJSON := originalJSON.(string)
	if srcJSON == "" {
		return nil
	}

	u, err := parseJSON(srcJSON)
	if err != nil {
		return logError(fmt.Errorf("JSON parse error: %s", err))
	}

	gvr, err := cgvk.getGVR(u.GroupVersionKind(), false)
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("failed to query GVR: %s", err),
		)
	}

	original, modified, current, err := getOriginalModifiedCurrent(
		originalJSON.(string),
		modifiedJSON.(string),
		true,
		m)
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("getOriginalModifiedCurrent failed: %s", err),
		)
	}

	patch, patchType, err := getPatch(u.GroupVersionKind(), original, modified, current)
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("getPatch failed: %s", err),
		)
	}

	dryRunPatch := k8smetav1.PatchOptions{DryRun: []string{k8smetav1.DryRunAll}}

	_, err = client.
		Resource(gvr).
		Namespace(u.GetNamespace()).
		Patch(context.TODO(), u.GetName(), patchType, patch, dryRunPatch)
	if err != nil {
		// Handle specific invalid errors
		if k8serrors.IsInvalid(err) {
			as := err.(k8serrors.APIStatus).Status()

			// ForceNew only when exact single cause
			if len(as.Details.Causes) == 1 {
				msg := as.Details.Causes[0].Message

				// if cause is immutable field force a delete and re-create plan
				if k8serrors.HasStatusCause(err, k8smetav1.CauseTypeFieldValueInvalid) && strings.HasSuffix(msg, ": field is immutable") == true {
					d.ForceNew("manifest")
					return nil
				}

				// if cause is statefulset forbidden fields error force a delete and re-create plan
				if k8serrors.HasStatusCause(err, k8smetav1.CauseType(field.ErrorTypeForbidden)) && strings.HasPrefix(msg, "Forbidden: updates to statefulset spec for fields") == true {
					d.ForceNew("manifest")
					return nil
				}
			}
		}

		return logErrorForResource(
			u,
			fmt.Errorf("patch failed '%s': %s", patchType, err),
		)
	}

	return nil
}

func kustomizationResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	srcJSON := d.Get("manifest").(string)
	u, err := parseJSON(srcJSON)
	if err != nil {
		return false, logError(fmt.Errorf("JSON parse error: %s", err))
	}

	gvr, err := cgvk.getGVR(u.GroupVersionKind(), false)
	if err != nil {
		if k8smeta.IsNoMatchError(err) {
			// If the Kind does not exist in the K8s API,
			// the resource can't exist either
			return false, nil
		}
		return false, err
	}

	_, err = client.
		Resource(gvr).
		Namespace(u.GetNamespace()).
		Get(context.TODO(), u.GetName(), k8smetav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, logErrorForResource(
			u,
			fmt.Errorf("get failed: %s", err),
		)
	}

	return true, nil
}

func kustomizationResourceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	originalJSON, modifiedJSON := d.GetChange("manifest")

	srcJSON := originalJSON.(string)
	u, err := parseJSON(srcJSON)
	if err != nil {
		return logError(fmt.Errorf("JSON parse error: %s", err))
	}

	if !d.HasChange("manifest") {
		return logErrorForResource(
			u,
			errors.New("update called without diff"),
		)
	}

	gvr, err := cgvk.getGVR(u.GroupVersionKind(), false)
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("failed to query GVR: %s", err),
		)
	}

	original, modified, current, err := getOriginalModifiedCurrent(
		originalJSON.(string),
		modifiedJSON.(string),
		false,
		m)
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("getOriginalModifiedCurrent failed: %s", err),
		)
	}

	patch, patchType, err := getPatch(u.GroupVersionKind(), original, modified, current)
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("getPatch failed: %s", err),
		)
	}

	var patchResp *unstructured.Unstructured
	patchResp, err = client.
		Resource(gvr).
		Namespace(u.GetNamespace()).
		Patch(context.TODO(), u.GetName(), patchType, patch, k8smetav1.PatchOptions{})
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("patch failed '%s': %s", patchType, err),
		)
	}

	id := string(patchResp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(patchResp))

	return kustomizationResourceRead(d, m)
}

func kustomizationResourceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	srcJSON := d.Get("manifest").(string)
	u, err := parseJSON(srcJSON)
	if err != nil {
		return logError(fmt.Errorf("JSON parse error: %s", err))
	}

	gvr, err := cgvk.getGVR(u.GroupVersionKind(), false)
	if err != nil {
		if k8smeta.IsNoMatchError(err) {
			// If the Kind does not exist in the K8s API,
			// the resource can't exist either
			return nil
		}
		return err
	}

	namespace := u.GetNamespace()
	name := u.GetName()

	err = client.
		Resource(gvr).
		Namespace(namespace).
		Delete(context.TODO(), name, k8smetav1.DeleteOptions{})
	if err != nil {
		// Consider not found during deletion a success
		if k8serrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}

		return logErrorForResource(
			u,
			fmt.Errorf("delete failed : %s", err),
		)
	}

	stateConf := &resource.StateChangeConf{
		Target:  []string{},
		Pending: []string{"deleting"},
		Timeout: d.Timeout(schema.TimeoutDelete),
		Refresh: func() (interface{}, string, error) {
			resp, err := client.
				Resource(gvr).
				Namespace(namespace).
				Get(context.TODO(), name, k8smetav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return nil, "", nil
				}
				return nil, "", err
			}

			return resp, "deleting", nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return logErrorForResource(
			u,
			fmt.Errorf("timed out waiting for delete: %s", err),
		)
	}

	d.SetId("")

	return nil
}

func kustomizationResourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	// "|" must match resid.separator
	if len(strings.Split(d.Id(), "|")) != 3 {
		return nil, logError(fmt.Errorf("invalid ID: %q, valid IDs look like: \"~G_v1_Namespace|~X|example\"", d.Id()))
	}

	rid := resid.FromString(d.Id())

	namespace := rid.Namespace
	name := rid.Name

	gvk := k8sschema.GroupVersionKind{
		Group:   rid.Gvk.Group,
		Version: rid.Gvk.Version,
		Kind:    rid.Gvk.Kind,
	}
	gvr, err := cgvk.getGVR(gvk, false)
	if err != nil {
		return nil, logError(
			fmt.Errorf("apiVersion: %q, kind: %q, namespace: %q, name: %q: failed to query GVR: %s", gvk.GroupVersion(), gvk.Kind, namespace, name, err),
		)
	}

	resp, err := client.
		Resource(gvr).
		Namespace(namespace).
		Get(context.TODO(), name, k8smetav1.GetOptions{})
	if err != nil {
		return nil, logError(
			fmt.Errorf("apiVersion: %q, kind: %q, namespace: %q, name: %q: get failed: %s", gvk.GroupVersion(), gvk.Kind, namespace, name, err),
		)
	}

	id := string(resp.GetUID())
	d.SetId(id)

	lac := getLastAppliedConfig(resp)
	if lac == "" {
		return nil, logError(
			fmt.Errorf("apiVersion: %q, kind: %q, namespace: %q, name: %q: can not import resources without %q annotation", gvk.GroupVersion(), gvk.Kind, namespace, name, lastAppliedConfig),
		)
	}

	d.Set("manifest", lac)

	return []*schema.ResourceData{d}, nil
}
