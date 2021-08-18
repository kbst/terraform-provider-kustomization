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
	"sigs.k8s.io/kustomize/kyaml/resid"
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

func runKustomizeBuild(fSys filesys.FileSystem, path string, kOpts *schema.ResourceData) (rm resmap.ResMap, err error) {

	opts := getKustomizeOptions(kOpts)

	k := krusty.MakeKustomizer(opts)

	rm, err = k.Run(fSys, path)
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

func getKustomizeOptions(d *schema.ResourceData) (opts *krusty.Options) {

	opts = krusty.MakeDefaultOptions()

	kOpts := d.Get("kustomize_options").(map[string]interface{})

	if kOpts["load_restrictor"] != nil {
		if kOpts["load_restrictor"].(string) == "none" {
			opts.LoadRestrictions = types.LoadRestrictionsNone
		}
	}

	return opts
}
