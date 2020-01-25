package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"sigs.k8s.io/kustomize/api/resid"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sunstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
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

func getGVR(u *k8sunstructured.Unstructured) k8sschema.GroupVersionResource {
	gvk := u.GroupVersionKind()

	return k8sschema.GroupVersionResource{
		Group:   gvk.Group,
		Version: gvk.Version,
		// TODO: fix this ugly shit
		Resource: strings.ToLower(gvk.Kind) + "s",
	}
}

func parseJSON(json string) (ur *k8sunstructured.Unstructured, err error) {
	body := []byte(json)
	u, err := k8sruntime.Decode(k8sunstructured.UnstructuredJSONScheme, body)
	if err != nil {
		return ur, err
	}

	ur = u.(*k8sunstructured.Unstructured)

	return ur, nil
}

func kustomizationResourceCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client

	srcJSON := d.Get("manifest").(string)
	u, err := parseJSON(srcJSON)
	if err != nil {
		return err
	}

	gvr := getGVR(u)
	namespace := u.GetNamespace()

	setLastAppliedConfig(u, srcJSON)

	resp, err := client.
		Resource(gvr).
		Namespace(namespace).
		Create(u, k8smetav1.CreateOptions{})
	if err != nil {
		return err
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(resp))

	return kustomizationResourceRead(d, m)
}

func kustomizationResourceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client

	u, err := parseJSON(d.Get("manifest").(string))
	if err != nil {
		return err
	}

	gvr := getGVR(u)
	namespace := u.GetNamespace()
	name := u.GetName()

	resp, err := client.
		Resource(gvr).
		Namespace(namespace).
		Get(name, k8smetav1.GetOptions{})
	if err != nil {
		return err
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(resp))

	return nil
}

func kustomizationResourceDiff(d *schema.ResourceDiff, m interface{}) error {
	client := m.(*Config).Client

	if !d.HasChange("manifest") {
		return nil
	}

	oldJSON, newJSON := d.GetChange("manifest")

	if oldJSON.(string) == "" {
		return nil
	}

	n, err := parseJSON(newJSON.(string))
	if err != nil {
		return err
	}
	o, err := parseJSON(oldJSON.(string))
	if err != nil {
		return err
	}

	gvr := getGVR(o)
	namespace := o.GetNamespace()
	name := o.GetName()

	setLastAppliedConfig(o, oldJSON.(string))
	setLastAppliedConfig(n, newJSON.(string))

	original, err := o.MarshalJSON()
	if err != nil {
		return err
	}

	modified, err := n.MarshalJSON()
	if err != nil {
		return err
	}

	c, err := client.
		Resource(gvr).
		Namespace(namespace).
		Get(name, k8smetav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	current, err := c.MarshalJSON()
	if err != nil {
		return err
	}

	patch, err := getPatch(original, modified, current)
	if err != nil {
		return err
	}

	dryRunPatch := k8smetav1.PatchOptions{DryRun: []string{k8smetav1.DryRunAll}}

	_, err = client.
		Resource(gvr).
		Namespace(namespace).
		Patch(name, k8stypes.MergePatchType, patch, dryRunPatch)
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
					return err
				}
			}

			d.ForceNew("manifest")
			return nil
		}
		return err
	}

	return nil
}

func kustomizationResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*Config).Client

	u, err := parseJSON(d.Get("manifest").(string))
	if err != nil {
		return false, err
	}

	gvr := getGVR(u)
	namespace := u.GetNamespace()
	name := u.GetName()

	_, err = client.
		Resource(gvr).
		Namespace(namespace).
		Get(name, k8smetav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func kustomizationResourceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client

	oldJSON, newJSON := d.GetChange("manifest")

	if !d.HasChange("manifest") {
		msg := fmt.Sprintf(
			"Update called without change. old: %s, new: %s",
			oldJSON,
			newJSON)
		return errors.New(msg)
	}

	n, err := parseJSON(newJSON.(string))
	if err != nil {
		return err
	}
	o, err := parseJSON(oldJSON.(string))
	if err != nil {
		return err
	}

	gvr := getGVR(o)
	namespace := o.GetNamespace()
	name := o.GetName()

	setLastAppliedConfig(o, oldJSON.(string))
	setLastAppliedConfig(n, newJSON.(string))

	original, err := o.MarshalJSON()
	if err != nil {
		return err
	}

	modified, err := n.MarshalJSON()
	if err != nil {
		return err
	}

	c, err := client.
		Resource(gvr).
		Namespace(namespace).
		Get(name, k8smetav1.GetOptions{})
	if err != nil {
		return err
	}

	current, err := c.MarshalJSON()
	if err != nil {
		return err
	}

	patch, err := getPatch(original, modified, current)
	if err != nil {
		return err
	}

	patchResp, err := client.
		Resource(gvr).
		Namespace(namespace).
		Patch(name, k8stypes.MergePatchType, patch, k8smetav1.PatchOptions{})
	if err != nil {
		return err
	}

	id := string(patchResp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(patchResp))

	return kustomizationResourceRead(d, m)
}

func kustomizationResourceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client

	u, err := parseJSON(d.Get("manifest").(string))
	if err != nil {
		return err
	}

	gvr := getGVR(u)
	namespace := u.GetNamespace()
	name := u.GetName()

	err = client.
		Resource(gvr).
		Namespace(namespace).
		Delete(name, nil)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Target:  []string{},
		Pending: []string{"deleting"},
		Timeout: d.Timeout(schema.TimeoutDelete),
		Refresh: func() (interface{}, string, error) {
			resp, err := client.
				Resource(gvr).
				Namespace(namespace).
				Get(name, k8smetav1.GetOptions{})
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
		return err
	}

	d.SetId("")

	return nil
}

func kustomizationResourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*Config).Client

	rid := resid.FromString(d.Id())

	gvr := k8sschema.GroupVersionResource{
		Group:   rid.Gvk.Group,
		Version: rid.Gvk.Version,
		// TODO: fix this ugly shit
		Resource: strings.ToLower(rid.Gvk.Kind) + "s",
	}
	namespace := rid.Namespace
	name := rid.Name

	resp, err := client.
		Resource(gvr).
		Namespace(namespace).
		Get(name, k8smetav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("importing '%s' failed: %s", d.Id(), err)
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(resp))

	return []*schema.ResourceData{d}, nil
}
