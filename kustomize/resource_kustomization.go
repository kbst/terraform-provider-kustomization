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
	k8stypes "k8s.io/apimachinery/pkg/types"
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
		return fmt.Errorf("ResourceCreate: JSON parse error: %s", err)
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
		return fmt.Errorf(
			"ResourceCreate: GroupVersionKind '%s' %s",
			u.GroupVersionKind(),
			err)
	}

	gvr := gvrResp.(k8sschema.GroupVersionResource)

	namespace := u.GetNamespace()
	name := u.GetName()

	setLastAppliedConfig(u, srcJSON)

	if namespace != "" {
		// wait for the namespace to exist
		nsGvk := k8sschema.GroupVersionKind{
			Group:   "",
			Version: "",
			Kind:    "Namespace"}
		nsGvr, err := cgvk.getGVR(nsGvk, false)
		if err != nil {
			return fmt.Errorf("ResourceCreate: '%s' %s/%s %s", gvr, namespace, name, err)
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
			return fmt.Errorf(
				"ResourceCreate: namespace '%s' %s",
				namespace,
				err)
		}
	}

	resp, err := client.
		Resource(gvr).
		Namespace(namespace).
		Create(context.TODO(), u, k8smetav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("ResourceCreate: creating '%s' failed: %s/%s %s", gvr, namespace, name, err)
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
		return fmt.Errorf("ResourceRead: JSON parse error: %s", err)
	}

	gvr, err := cgvk.getGVR(u.GroupVersionKind(), false)
	if err != nil {
		return fmt.Errorf("ResourceRead: %s", err)
	}

	namespace := u.GetNamespace()
	name := u.GetName()

	resp, err := client.
		Resource(gvr).
		Namespace(namespace).
		Get(context.TODO(), name, k8smetav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("ResourceRead: reading '%s' failed: %s/%s %s", gvr, namespace, name, err)
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
		return fmt.Errorf("ResourceDiff: JSON parse error: %s", err)
	}

	gvr, err := cgvk.getGVR(u.GroupVersionKind(), false)
	if err != nil {
		return fmt.Errorf("ResourceDiff: %s", err)
	}

	namespace := u.GetNamespace()
	name := u.GetName()

	original, modified, current, err := getOriginalModifiedCurrent(
		originalJSON.(string),
		modifiedJSON.(string),
		true,
		m)
	if err != nil {
		return fmt.Errorf("ResourceDiff: '%s' %s/%s %s", gvr, namespace, name, err)
	}

	patch, err := getPatch(original, modified, current)
	if err != nil {
		return fmt.Errorf("ResourceDiff: '%s' %s/%s %s", gvr, namespace, name, err)
	}

	dryRunPatch := k8smetav1.PatchOptions{DryRun: []string{k8smetav1.DryRunAll}}

	patchTypes := []k8stypes.PatchType{
		k8stypes.StrategicMergePatchType,
		k8stypes.MergePatchType,
	}
	for _, patchType := range patchTypes {
		_, err = client.
			Resource(gvr).
			Namespace(namespace).
			Patch(context.TODO(), name, patchType, patch, dryRunPatch)
		if err != nil {

			//
			// If the resource kind does not support StrategicMergePatchType
			// fall back to MergePatchType and retry
			if k8serrors.IsUnsupportedMediaType(err) {
				continue
			}

			//
			// Find out if the request is invalid because a field is immutable
			// if immutable is the only reason, force a delete and recreate plan
			if k8serrors.IsInvalid(err) {
				as := err.(k8serrors.APIStatus).Status()

				for _, c := range as.Details.Causes {
					if strings.HasSuffix(c.Message, ": field is immutable") != true {
						// if there is any error that is not due to an immutable field
						// expose to user to let them fix it first
						return fmt.Errorf("ResourceDiff: '%s' %s/%s %s", gvr, namespace, name, err)
					}
				}

				d.ForceNew("manifest")
				return nil
			}

			return fmt.Errorf("ResourceDiff: '%s' %s/%s %s", gvr, namespace, name, err)
		}

		// If StrategicMergePatchType succeeded without error stop the loop
		break
	}

	return nil
}

func kustomizationResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	srcJSON := d.Get("manifest").(string)
	u, err := parseJSON(srcJSON)
	if err != nil {
		return false, fmt.Errorf("ResourceExists: JSON parse error: %s", err)
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
	namespace := u.GetNamespace()
	name := u.GetName()

	_, err = client.
		Resource(gvr).
		Namespace(namespace).
		Get(context.TODO(), name, k8smetav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("ResourceExists: reading '%s' failed: %s", gvr, err)
	}

	return true, nil
}

func kustomizationResourceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	originalJSON, modifiedJSON := d.GetChange("manifest")

	if !d.HasChange("manifest") {
		msg := fmt.Sprintf(
			"Update called without change. old: %s, new: %s",
			originalJSON,
			modifiedJSON)
		return errors.New(msg)
	}

	srcJSON := originalJSON.(string)
	u, err := parseJSON(srcJSON)
	if err != nil {
		return fmt.Errorf("ResourceUpdate: JSON parse error: %s", err)
	}

	gvr, err := cgvk.getGVR(u.GroupVersionKind(), false)
	if err != nil {
		return fmt.Errorf("ResourceUpdate: %s", err)
	}

	namespace := u.GetNamespace()
	name := u.GetName()

	original, modified, current, err := getOriginalModifiedCurrent(
		originalJSON.(string),
		modifiedJSON.(string),
		false,
		m)
	if err != nil {
		return fmt.Errorf("ResourceUpdate: '%s' %s/%s %s", gvr, namespace, name, err)
	}

	patch, err := getPatch(original, modified, current)
	if err != nil {
		return fmt.Errorf("ResourceUpdate: '%s' %s/%s %s", gvr, namespace, name, err)
	}

	var patchResp *unstructured.Unstructured
	patchTypes := []k8stypes.PatchType{
		k8stypes.StrategicMergePatchType,
		k8stypes.MergePatchType,
	}
	for _, patchType := range patchTypes {
		patchResp, err = client.
			Resource(gvr).
			Namespace(namespace).
			Patch(context.TODO(), name, patchType, patch, k8smetav1.PatchOptions{})
		if err != nil {
			//
			// If the resource kind does not support StrategicMergePatchType
			// fall back to MergePatchType and retry
			if k8serrors.IsUnsupportedMediaType(err) {
				continue
			}

			return fmt.Errorf("ResourceUpdate: patching '%s' failed: %s/%s %s", gvr, namespace, name, err)
		}

		// If StrategicMergePatchType succeeded without error stop the loop
		break
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
		return fmt.Errorf("ResourceDelete: JSON parse error: %s", err)
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

		return fmt.Errorf("ResourceDelete: deleting '%s' failed: %s/%s %s", gvr, namespace, name, err)
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
				return nil, "", fmt.Errorf("refreshing '%s' state failed: %s/%s %s", gvr, namespace, name, err)
			}

			return resp, "deleting", nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("ResourceDelete: '%s' %s/%s %s", gvr, namespace, name, err)
	}

	d.SetId("")

	return nil
}

func kustomizationResourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*Config).Client
	cgvk := m.(*Config).CachedGroupVersionKind

	rid := resid.FromString(d.Id())

	gvk := k8sschema.GroupVersionKind{
		Group:   rid.Gvk.Group,
		Version: rid.Gvk.Version,
		Kind:    rid.Gvk.Kind,
	}
	gvr, err := cgvk.getGVR(gvk, false)
	if err != nil {
		return nil, fmt.Errorf("ResourceImport: %s", err)
	}

	namespace := rid.Namespace
	name := rid.Name

	resp, err := client.
		Resource(gvr).
		Namespace(namespace).
		Get(context.TODO(), name, k8smetav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("ResourceImport: reading '%s' failed: %s", gvr, err)
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(resp))

	return []*schema.ResourceData{d}, nil
}
