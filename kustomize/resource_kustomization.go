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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const fieldManager string = "terraform-provider-kustomization"

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
			"wait": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func kustomizationResourceCreate(d *schema.ResourceData, m interface{}) error {
	mapper := m.(*Config).Mapper
	client := m.(*Config).Client
	extractor := m.(*Config).Extractor
	gzipLastAppliedConfig := m.(*Config).GzipLastAppliedConfig
	serverSideApply := m.(*Config).ServerSideApply
	serverSideApplyForce := m.(*Config).ServerSideApplyForce

	km := newKManifest(mapper, client, extractor)

	err := km.load([]byte(d.Get("manifest").(string)))
	if err != nil {
		return logError(err)
	}

	// required for CRDs
	err = km.waitKind(d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return logError(err)
	}

	// required for namespaced resources
	err = km.waitNamespace(d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return logError(err)
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
					return logError(km.fmtErr(
						fmt.Errorf("api error: %q: %s", saGvk.String(), err),
					))
				}

				_, err = waitForGVKCreated(d, client, mapping, km.namespace(), v)
				if err != nil {
					return logError(km.fmtErr(fmt.Errorf("timed out waiting for: %q: %s", km.id().string(), err)))
				}
			}
		}
	}

	var resp *unstructured.Unstructured

	if serverSideApply {
		resp, err = km.apiApply(k8smetav1.ApplyOptions{
			FieldManager: fieldManager,
			Force:        serverSideApplyForce,
		})
		if err != nil {
			return logError(fmt.Errorf("server-side apply error: %s", err))
		}
	}

	// if server-side apply disabled or failed, fallback to create
	if !serverSideApply {
		setLastAppliedConfig(km, gzipLastAppliedConfig)

		resp, err = km.apiCreate(k8smetav1.CreateOptions{
			FieldManager: fieldManager,
		})
		if err != nil {
			return logError(fmt.Errorf("create failed: %s", err))
		}
	}

	if d.Get("wait").(bool) {
		if err = km.waitCreatedOrUpdated(d.Timeout(schema.TimeoutCreate)); err != nil {
			return logError(err)
		}
	}

	id := string(resp.GetUID())
	d.SetId(id)

	lac := extractLastAppliedConfig(resp, extractor, gzipLastAppliedConfig)
	d.Set("manifest", lac)

	return nil
}

func kustomizationResourceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	mapper := m.(*Config).Mapper
	extractor := m.(*Config).Extractor
	gzipLastAppliedConfig := m.(*Config).GzipLastAppliedConfig

	km := newKManifest(mapper, client, extractor)

	err := km.load([]byte(d.Get("manifest").(string)))
	if err != nil {
		return logError(err)
	}

	resp, err := km.apiGet(k8smetav1.GetOptions{})
	if err != nil {
		return logError(err)
	}

	id := string(resp.GetUID())
	d.SetId(id)

	lac := extractLastAppliedConfig(resp, extractor, gzipLastAppliedConfig)
	d.Set("manifest", lac)

	return nil
}

func kustomizationResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*Config).Client
	mapper := m.(*Config).Mapper
	extractor := m.(*Config).Extractor

	km := newKManifest(mapper, client, extractor)

	err := km.load([]byte(d.Get("manifest").(string)))
	if err != nil {
		return false, logError(err)
	}

	err = km.waitKind(d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if k8smeta.IsNoMatchError(err) {
			// If the Kind does not exist in the K8s API,
			// the resource can't exist either
			return false, nil
		}
		return false, logError(err)
	}

	_, err = km.apiGet(k8smetav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, logError(err)
	}

	return true, nil
}

func kustomizationResourceDiff(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	if !d.HasChange("manifest") {
		return nil
	}

	client := m.(*Config).Client
	mapper := m.(*Config).Mapper
	extractor := m.(*Config).Extractor
	gzipLastAppliedConfig := m.(*Config).GzipLastAppliedConfig
	serverSideApply := m.(*Config).ServerSideApply
	serverSideApplyForce := m.(*Config).ServerSideApplyForce

	do, dm := d.GetChange("manifest")

	kmm := newKManifest(mapper, client, extractor)
	err := kmm.load([]byte(dm.(string)))
	if err != nil {
		return logError(err)
	}
	setLastAppliedConfig(kmm, gzipLastAppliedConfig)

	_, err = kmm.mappings()
	if err != nil {
		// if there are no mappings we can't dry-run
		// this is for CRDs that do not exist yet
		return nil
	}

	isNamespaced, err := kmm.isNamespaced()
	if err != nil {
		return logError(err)
	}
	if isNamespaced && kmm.namespace() == "" {
		err = kmm.fmtErr(fmt.Errorf("is namespace scoped and must set metadata.namespace"))
		return logError(err)
	}
	if !isNamespaced && kmm.namespace() != "" {
		err = kmm.fmtErr(fmt.Errorf("is not namespace scoped but has metadata.namespace set"))
		return logError(err)
	}

	if do.(string) == "" {
		_, err = kmm.apiCreate(k8smetav1.CreateOptions{
			FieldManager: fieldManager,
		})

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

			return logError(kmm.fmtErr(err))
		}

		return nil
	}

	kmo := newKManifest(mapper, client, extractor)
	err = kmo.load([]byte(do.(string)))
	if err != nil {
		return logError(err)
	}
	setLastAppliedConfig(kmo, gzipLastAppliedConfig)

	if kmo.name() != kmm.name() || kmo.namespace() != kmm.namespace() {
		// if the resource name or namespace changes, we can't patch but have to destroy and re-create
		d.ForceNew("manifest")
		return nil
	}

	if serverSideApply {
		_, err = kmm.apiApply(k8smetav1.ApplyOptions{
			DryRun:       []string{k8smetav1.DryRunAll},
			FieldManager: fieldManager,
			Force:        serverSideApplyForce,
		})
	}

	// if server-side apply disabled fallback to patch
	if !serverSideApply {
		kmo := newKManifest(mapper, client, extractor)
		err = kmo.load([]byte(do.(string)))
		if err != nil {
			return logError(err)
		}
		setLastAppliedConfig(kmo, gzipLastAppliedConfig)

		pt, p, pErr := kmm.apiPreparePatch(kmo, true)
		if pErr != nil {
			return logError(pErr)
		}

		_, err = kmm.apiPatch(pt, p, k8smetav1.PatchOptions{
			DryRun:       []string{k8smetav1.DryRunAll},
			FieldManager: fieldManager,
		})
	}

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

				// if cause is updates to storage class provisioner or parameters are forbidden force a delete and re-create plan
				if k8serrors.HasStatusCause(err, k8smetav1.CauseType(field.ErrorTypeForbidden)) {
					if strings.HasSuffix(msg, ": updates to provisioner are forbidden.") || strings.HasPrefix(msg, "Forbidden: updates to parameters are forbidden") {
						d.ForceNew("manifest")
						return nil
					}
				}

			}
		}

		return logError(err)
	}

	return nil
}

func kustomizationResourceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	mapper := m.(*Config).Mapper
	extractor := m.(*Config).Extractor
	gzipLastAppliedConfig := m.(*Config).GzipLastAppliedConfig
	serverSideApply := m.(*Config).ServerSideApply
	serverSideApplyForce := m.(*Config).ServerSideApplyForce

	do, dm := d.GetChange("manifest")

	kmm := newKManifest(mapper, client, extractor)
	err := kmm.load([]byte(dm.(string)))
	if err != nil {
		return logError(err)
	}
	setLastAppliedConfig(kmm, gzipLastAppliedConfig)

	if !d.HasChange("manifest") && !d.HasChange("wait") {
		return logError(kmm.fmtErr(
			errors.New("update called without diff"),
		))
	}

	var resp *unstructured.Unstructured
	if serverSideApply {
		resp, err = kmm.apiApply(k8smetav1.ApplyOptions{
			FieldManager: fieldManager,
			Force:        serverSideApplyForce,
		})
		if err != nil {
			logError(fmt.Errorf("server-side apply error: %s", err))
		}
	}

	// if server-side apply disabled fallback to patch
	if !serverSideApply {
		kmo := newKManifest(mapper, client, extractor)
		err = kmo.load([]byte(do.(string)))
		if err != nil {
			return logError(err)
		}
		setLastAppliedConfig(kmo, gzipLastAppliedConfig)

		pt, p, err := kmm.apiPreparePatch(kmo, true)
		if err != nil {
			return logError(err)
		}

		resp, err = kmm.apiPatch(pt, p, k8smetav1.PatchOptions{
			FieldManager: fieldManager,
		})
		if err != nil {
			return logError(err)
		}
	}

	if d.Get("wait").(bool) {
		if err = kmm.waitCreatedOrUpdated(d.Timeout(schema.TimeoutUpdate)); err != nil {
			return logError(err)
		}
	}

	id := string(resp.GetUID())
	d.SetId(id)

	lac := extractLastAppliedConfig(resp, extractor, gzipLastAppliedConfig)
	d.Set("manifest", lac)

	return nil
}

func kustomizationResourceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Config).Client
	mapper := m.(*Config).Mapper
	extractor := m.(*Config).Extractor

	km := newKManifest(mapper, client, extractor)

	err := parseResourceData(km, d.Get("manifest").(string))
	if err != nil {
		return logError(err)
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
		return logError(km.fmtErr(err))
	}

	err = km.apiDelete(k8smetav1.DeleteOptions{})
	if err != nil {
		// Consider not found during deletion a success
		if k8serrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}

		return logError(err)
	}

	err = km.waitDeleted(d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return logError(err)
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
	d.Set("wait", d.Get("wait"))

	return []*schema.ResourceData{d}, nil
}
