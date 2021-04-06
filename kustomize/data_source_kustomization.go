package kustomize

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"log"
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

func determinePrefix(gvk resid.Gvk) (p uint32) {
	// Default prefix to 5
	p = 5

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

	return p
}

func prefixHash(p uint32, h uint32) int {
	s := fmt.Sprintf("%01d%010d", p, h)
	s = s[0:9]

	i, e := strconv.ParseInt(s, 10, 32)
	if e != nil {
		// return unmodified hash
		log.Printf("idSetHash: %s", e)
		return int(h)
	}

	return int(i)
}

func idSetHash(v interface{}) int {
	id := v.(string)

	gvk := resid.GvkFromString(id)
	p := determinePrefix(gvk)
	h := crc32.ChecksumIEEE([]byte(id))

	return prefixHash(p, h)
}

func runKustomizeBuild(fSys filesys.FileSystem, path string, load_restrictor string) (rm resmap.ResMap, err error) {
	opts := krusty.MakeDefaultOptions()

	// If the load_restrictor option is set to none then the krusty.Options
	// should be updated to reflect this.
	if load_restrictor == "none" {
		opts.LoadRestrictions = types.LoadRestrictionsNone
	}

	k := krusty.MakeKustomizer(fSys, opts)

	rm, err = k.Run(path)
	if err != nil {
		return nil, fmt.Errorf("Kustomizer Run for path '%s' failed: %s", path, err)
	}

	return rm, nil
}

func setGeneratedAttributes(d *schema.ResourceData, rm resmap.ResMap) error {
	ids, idsPrio := flattenKustomizationIDs(rm)
	d.Set("ids", ids)
	d.Set("ids_prio", idsPrio)

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

type kustomizeOptions struct {
	loadRestrictor string
}

func getKustomizeOptions(d *schema.ResourceData) (k kustomizeOptions) {
	// initialize kustomizeOptions with defaults
	k = kustomizeOptions{
		loadRestrictor: "",
	}

	kOpts := d.Get("kustomize_options").(map[string]interface{})

	if kOpts["load_restrictor"] != nil {
		if kOpts["load_restrictor"].(string) == "none" {
			k.loadRestrictor = "none"
		}
	}

	return k
}
