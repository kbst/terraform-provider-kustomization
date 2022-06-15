package kustomize

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smeta "k8s.io/apimachinery/pkg/api/meta"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	mapper := m.(*Config).Mapper
	client := m.(*Config).Client
	km := newKManifest(mapper, client)

	err := km.load([]byte(d.Get("manifest").(string)))
	if err != nil {
		return err
	}

	// required for CRDs
	err = km.waitKind(time.Duration(3 * time.Minute))
	if err != nil {
		return err
	}

	// required for namespaced resources
	err = km.waitNamespace(time.Duration(5 * time.Minute))
	if err != nil {
		return err
	}

	// for secrets of type service account token
	// wait for service account to exist
	// https://github.com/kubernetes/kubernetes/issues/109401
	if (km.gvk().Kind == "Secret") &&
		(km.resource.UnstructuredContent()["type"] != nil) &&
		(km.resource.UnstructuredContent()["type"].(string) == string(k8scorev1.SecretTypeServiceAccountToken)) {

		annotations := km.resource.GetAnnotations()
		for k, v := range annotations {
			if k == k8scorev1.ServiceAccountNameKey {
				saGvk := k8sschema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "ServiceAccount"}
				mapping, err := mapper.RESTMapping(saGvk.GroupKind(), saGvk.GroupVersion().Version)
				if err != nil {
					return km.fmtErr(
						fmt.Errorf("api error: %q: %s", saGvk.String(), err),
					)
				}

				_, err = waitForGVKCreated(d, client, mapping, km.namespace(), v)
				if err != nil {
					return km.fmtErr(fmt.Errorf("timed out waiting for: %q: %s", km.id().toString(), err))
				}
			}
		}
	}

	gzipLastAppliedConfig := m.(*Config).GzipLastAppliedConfig
	setLastAppliedConfig(km, gzipLastAppliedConfig)

	resp, err := km.apiCreate(k8smetav1.CreateOptions{})
	if err != nil {
		return err
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(resp, gzipLastAppliedConfig))

	return kustomizationResourceRead(d, m)
}

func kustomizationResourceRead(d *schema.ResourceData, m interface{}) error {
	km := newKManifest(m.(*Config).Mapper, m.(*Config).Client)

	err := km.load([]byte(d.Get("manifest").(string)))
	if err != nil {
		return err
	}

	resp, err := km.apiGet(k8smetav1.GetOptions{})
	if err != nil {
		return err
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(resp, m.(*Config).GzipLastAppliedConfig))

	return nil
}

func kustomizationResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	km := newKManifest(m.(*Config).Mapper, m.(*Config).Client)

	err := km.load([]byte(d.Get("manifest").(string)))
	if err != nil {
		return false, err
	}

	err = km.waitKind(time.Duration(5 * time.Second))
	if err != nil {
		if k8smeta.IsNoMatchError(err) {
			// If the Kind does not exist in the K8s API,
			// the resource can't exist either
			return false, nil
		}
		return false, err
	}

	_, err = km.apiGet(k8smetav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func kustomizationResourceDiff(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	if !d.HasChange("manifest") {
		return nil
	}

	client := m.(*Config).Client
	mapper := m.(*Config).Mapper
	gzipLastAppliedConfig := m.(*Config).GzipLastAppliedConfig

	do, dm := d.GetChange("manifest")

	kmm := newKManifest(mapper, client)
	err := kmm.load([]byte(dm.(string)))
	if err != nil {
		return err
	}
	setLastAppliedConfig(kmm, gzipLastAppliedConfig)

	if do.(string) == "" {
		// diffing for create
		_, err := kmm.mappings()
		if err != nil {
			// if there are no mappings we can't dry-run
			// this is for CRDs that do not exist yet
			return nil
		}

		_, err = kmm.apiCreate(k8smetav1.CreateOptions{DryRun: []string{k8smetav1.DryRunAll}})
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				// this is an edge case during tests
				// get change above has empty original
				// yet the create request fails with
				// Error running pre-apply refresh
				return nil
			}

			if k8serrors.IsNotFound(err) {
				// we're dry-running a create
				// the notfound seems mostly the namespace
				return nil
			}

			return kmm.fmtErr(err)
		}

		return nil
	}

	// diffing for update
	kmo := newKManifest(mapper, client)
	err = kmo.load([]byte(do.(string)))
	if err != nil {
		return err
	}
	setLastAppliedConfig(kmo, gzipLastAppliedConfig)

	if kmo.name() != kmm.name() || kmo.namespace() != kmm.namespace() {
		// if the resource name or namespace changes, we can't patch but have to destroy and re-create
		d.ForceNew("manifest")
		return nil
	}

	pt, p, err := kmm.apiPreparePatch(kmo, true)
	if err != nil {
		return err
	}

	dryRunPatch := k8smetav1.PatchOptions{DryRun: []string{k8smetav1.DryRunAll}}

	_, err = kmm.apiPatch(pt, p, dryRunPatch)
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

				// if cause is cannot change roleRef force a delete and re-create plan
				if k8serrors.HasStatusCause(err, k8smetav1.CauseTypeFieldValueInvalid) && strings.HasSuffix(msg, ": cannot change roleRef") == true {
					d.ForceNew("manifest")
					return nil
				}

			}
		}

		return err
	}

	return nil
}

func kustomizationResourceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	mapper := m.(*Config).Mapper
	gzipLastAppliedConfig := m.(*Config).GzipLastAppliedConfig

	do, dm := d.GetChange("manifest")

	kmo := newKManifest(mapper, client)
	err := kmo.load([]byte(do.(string)))
	if err != nil {
		return err
	}

	kmm := newKManifest(mapper, client)
	err = kmm.load([]byte(dm.(string)))
	if err != nil {
		return err
	}

	if !d.HasChange("manifest") {
		return kmm.fmtErr(
			errors.New("update called without diff"),
		)
	}

	setLastAppliedConfig(kmo, gzipLastAppliedConfig)
	setLastAppliedConfig(kmm, gzipLastAppliedConfig)

	pt, p, err := kmm.apiPreparePatch(kmo, false)
	if err != nil {
		return err
	}

	resp, err := kmm.apiPatch(pt, p, k8smetav1.PatchOptions{})
	if err != nil {
		return err
	}

	id := string(resp.GetUID())
	d.SetId(id)

	d.Set("manifest", getLastAppliedConfig(resp, gzipLastAppliedConfig))

	return kustomizationResourceRead(d, m)
}

func kustomizationResourceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	mapper := m.(*Config).Mapper

	km := newKManifest(mapper, client)

	err := parseResourceData(km, d.Get("manifest").(string))
	if err != nil {
		return err
	}

	// look for all versions of the GroupKind in case the resource uses a
	// version that is no longer current
	_, err = km.mappings()
	if err != nil {
		if k8smeta.IsNoMatchError(err) {
			// If the Kind does not exist in the K8s API,
			// the resource can't exist either
			return nil
		}
		return err
	}

	err = km.apiDelete(k8smetav1.DeleteOptions{})
	if err != nil {
		// Consider not found during deletion a success
		if k8serrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}

		return err
	}

	err = km.waitDeleted(time.Duration(7 * time.Minute))
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func kustomizationResourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*Config).Client
	mapper := m.(*Config).Mapper
	gzipLastAppliedConfig := m.(*Config).GzipLastAppliedConfig

	k, err := parseProviderId(d.Id())
	if err != nil {
		return nil, logError(err)
	}
	gk := k8sschema.GroupKind{Group: k.group, Kind: k.kind}

	// We don't need to use a specific API version here, as we're going to store the
	// resource using the LastAppliedConfig information which we can get from any
	// API version
	mappings, err := mapper.RESTMappings(gk)
	if err != nil {
		return nil, logError(
			fmt.Errorf("api error \"%s/%s/%s/%s\": %s", gk.Group, gk.Kind, k.namespace, k.name, err),
		)
	}

	resp, err := client.
		Resource(mappings[0].Resource).
		Namespace(k.namespace).
		Get(context.TODO(), k.name, k8smetav1.GetOptions{})
	if err != nil {
		return nil, logError(
			fmt.Errorf("\"%s/%s/%s/%s\": %s", gk.Group, gk.Kind, k.namespace, k.name, err),
		)
	}

	id := string(resp.GetUID())
	d.SetId(id)

	lac := getLastAppliedConfig(resp, gzipLastAppliedConfig)
	if lac == "" {
		return nil, logError(
			fmt.Errorf("\"%s/%s/%s/%s\": can not import resources without %q or %q annotation", gk.Group, gk.Kind, k.namespace, k.name, lastAppliedConfigAnnotation, gzipLastAppliedConfigAnnotation),
		)
	}

	d.Set("manifest", lac)

	return []*schema.ResourceData{d}, nil
}
