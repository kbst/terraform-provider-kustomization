package kustomize

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
)

func getIDFromResources(rm resmap.ResMap) (s string, err error) {
	h := sha512.New()

	yaml, err := rm.AsYaml()
	if err != nil {
		return "", fmt.Errorf("ResMap AsYaml failed: %s", err)
	}
	h.Write(yaml)

	s = hex.EncodeToString(h.Sum(nil))

	return s, nil
}

func dataSourceKustomization() *schema.Resource {
	return &schema.Resource{
		Read: kustomizationBuild,

		Schema: map[string]*schema.Schema{
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"ids": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      idSetHash,
			},
			"manifests": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func idSetHash(v interface{}) int {
	id := v.(string)

	h := crc32.ChecksumIEEE([]byte(id))
	f := int(h)

	gvk := resid.GvkFromString(id)

	// Default prefix to 5
	var p uint32 = 5

	for _, k := range []string{
		"Namespace",
		"CustomResourceDefinition",
	} {
		if strings.HasPrefix(gvk.Kind, k) {
			p = 1
		}
	}

	for _, k := range []string{
		"MutatingWebhookConfiguration",
		"ValidatingWebhookConfiguration",
	} {
		if strings.HasPrefix(gvk.Kind, k) {
			p = 9
		}
	}

	s := fmt.Sprintf("%01d%010d", p, h)

	i, e := strconv.ParseInt(s, 10, 32)
	if e != nil {
		// return unmodified hash
		return f
	}
	f = int(i)

	return f
}

func runKustomizeBuild(path string) (rm resmap.ResMap, err error) {
	fSys := filesys.MakeFsOnDisk()
	opts := &krusty.Options{
		DoLegacyResourceSort: true,
		LoadRestrictions:     types.LoadRestrictionsRootOnly,
		DoPrune:              false,
	}

	k := krusty.MakeKustomizer(fSys, opts)

	rm, err = k.Run(path)
	if err != nil {
		return nil, fmt.Errorf("Kustomizer Run for path '%s' failed: %s", path, err)
	}

	return rm, nil
}

func kustomizationBuild(d *schema.ResourceData, m interface{}) error {
	path := d.Get("path").(string)
	rm, err := runKustomizeBuild(path)
	if err != nil {
		return fmt.Errorf("kustomizationBuild: %s", err)
	}

	d.Set("ids", flattenKustomizationIDs(rm))

	resources, err := flattenKustomizationResources(rm)
	if err != nil {
		return fmt.Errorf("kustomizationBuild: %s", err)
	}
	d.Set("manifests", resources)

	id, err := getIDFromResources(rm)
	if err != nil {
		return fmt.Errorf("kustomizationBuild: %s", err)
	}
	d.SetId(id)

	return nil
}
