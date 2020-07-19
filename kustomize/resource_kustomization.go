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
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	clientset := m.(*Config).Clientset

	srcJSON := d.Get("manifest").(string)
	u, err := parseJSON(srcJSON)
	if err != nil {
		return fmt.Errorf("ResourceCreate: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Target:  []string{"existing"},
		Pending: []string{"pending"},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: func() (interface{}, string, error) {
			// CRDs: wait for GroupVersionKind to exist
			gvr, err := getGVR(u.GroupVersionKind(), clientset)
			if err != nil {
				return nil, "pending", nil
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

	if namespace != "" {
		// wait for the namespace to exist
		nsGvk := k8sschema.GroupVersionKind{
			Group:   "",
			Version: "",
			Kind:    "Namespace"}
		nsGvr, err := getGVR(nsGvk, clientset)
		if err != nil {
			return fmt.Errorf("ResourceCreate: %s", err)
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
		return fmt.Errorf("ResourceCreate: creating '%s' failed: %s", gvr, err)
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getSimplified(resp, []byte(srcJSON)))

	return kustomizationResourceRead(d, m)
}

func kustomizationResourceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	clientset := m.(*Config).Clientset

	srcJSON := d.Get("manifest").(string)
	u, err := parseJSON(srcJSON)

	if err != nil {
		return fmt.Errorf("ResourceRead: %s", err)
	}

	gvr, err := getGVR(u.GroupVersionKind(), clientset)
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
		return fmt.Errorf("ResourceRead: reading '%s' failed: %s", gvr, err)
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getSimplified(resp, []byte(srcJSON)))

	return nil
}

func kustomizationResourceDiff(d *schema.ResourceDiff, m interface{}) error {
	client := m.(*Config).Client
	clientset := m.(*Config).Clientset

	originalJSON, modifiedJSON := d.GetChange("manifest")

	if !d.HasChange("manifest") {
		return nil
	}

	if originalJSON.(string) == "" {
		return nil
	}

	u, err := parseJSON(originalJSON.(string))
	if err != nil {
		return fmt.Errorf("ResourceDiff: %s", err)
	}

	gvr, err := getGVR(u.GroupVersionKind(), clientset)
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
		return fmt.Errorf("ResourceDiff: %s", err)
	}

	patch, err := getPatch(original, modified, current)
	if err != nil {
		return fmt.Errorf("ResourceDiff: %s", err)
	}

	dryRunPatch := k8smetav1.PatchOptions{DryRun: []string{k8smetav1.DryRunAll}}

	_, err = client.
		Resource(gvr).
		Namespace(namespace).
		Patch(context.TODO(), name, k8stypes.StrategicMergePatchType, patch, dryRunPatch)
	if err != nil {
		//
		//
		// Find out if the request is invalid because a field is immutable
		// if immutable is the only reason, force a delete and recreate plan
		if k8serrors.IsInvalid(err) {
			as := err.(k8serrors.APIStatus).Status()

			for _, c := range as.Details.Causes {
				if strings.HasSuffix(c.Message, ": field is immutable") != true {
					// if there is any error that is not due to an immutable field
					// expose to user to let them fix it first
					return fmt.Errorf("ResourceDiff: %s", err)
				}
			}

			d.ForceNew("manifest")
			return nil
		}
		return fmt.Errorf("ResourceDiff: %s", err)
	}

	return nil
}

func kustomizationResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*Config).Client
	clientset := m.(*Config).Clientset

	u, err := parseJSON(d.Get("manifest").(string))
	if err != nil {
		return false, fmt.Errorf("ResourceExists: %s", err)
	}

	gvr, err := getGVR(u.GroupVersionKind(), clientset)
	if err != nil {
		return false, fmt.Errorf("ResourceExists: %s", err)
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
	clientset := m.(*Config).Clientset

	originalJSON, modifiedJSON := d.GetChange("manifest")

	if !d.HasChange("manifest") {
		msg := fmt.Sprintf(
			"Update called without change. old: %s, new: %s",
			originalJSON,
			modifiedJSON)
		return errors.New(msg)
	}

	u, err := parseJSON(originalJSON.(string))
	if err != nil {
		return fmt.Errorf("ResourceUpdate: %s", err)
	}

	gvr, err := getGVR(u.GroupVersionKind(), clientset)
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
		return fmt.Errorf("ResourceUpdate: %s", err)
	}

	patch, err := getPatch(original, modified, current)
	if err != nil {
		return fmt.Errorf("ResourceUpdate: %s", err)
	}

	patchResp, err := client.
		Resource(gvr).
		Namespace(namespace).
		Patch(context.TODO(), name, k8stypes.StrategicMergePatchType, patch, k8smetav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("ResourceUpdate: patching '%s' failed: %s", gvr, err)
	}

	id := string(patchResp.GetUID())
	d.SetId(id)

	d.Set("manifest", getSimplified(patchResp, modified))

	return kustomizationResourceRead(d, m)
}

func kustomizationResourceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	clientset := m.(*Config).Clientset

	u, err := parseJSON(d.Get("manifest").(string))
	if err != nil {
		return fmt.Errorf("ResourceDelete: %s", err)
	}

	gvr, err := getGVR(u.GroupVersionKind(), clientset)
	if err != nil {
		return fmt.Errorf("ResourceDelete: %s", err)
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

		return fmt.Errorf("ResourceDelete: deleting '%s' failed: %s", gvr, err)
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
				return nil, "", fmt.Errorf("refreshing '%s' state failed: %s", gvr, err)
			}

			return resp, "deleting", nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("ResourceDelete: %s", err)
	}

	d.SetId("")

	return nil
}

func kustomizationResourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*Config).Client
	clientset := m.(*Config).Clientset

	rid := resid.FromString(d.Id())

	gvk := k8sschema.GroupVersionKind{
		Group:   rid.Gvk.Group,
		Version: rid.Gvk.Version,
		Kind:    rid.Gvk.Kind,
	}
	gvr, err := getGVR(gvk, clientset)
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

	srcJSON := d.Get("manifest").(string)
	d.Set("manifest", getSimplified(resp, []byte(srcJSON)))

	return []*schema.ResourceData{d}, nil
}
