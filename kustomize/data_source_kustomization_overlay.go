package kustomize

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func dataSourceKustomizationOverlay() *schema.Resource {
	return &schema.Resource{
		Read: kustomizationOverlay,

		Schema: map[string]*schema.Schema{
			"namespace": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"resources": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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

func buildKustomizeOverlay(kustomization types.Kustomization) (rm resmap.ResMap, err error) {
	fSys := filesys.MakeFsOnDisk()
	opts := krusty.MakeDefaultOptions()

	var b bytes.Buffer
	ye := yaml.NewEncoder(io.Writer(&b))
	ye.Encode(kustomization)
	ye.Close()
	data, _ := ioutil.ReadAll(io.Reader(&b))

	fSys.WriteFile("Kustomization", data)
	defer fSys.RemoveAll("Kustomization")

	k := krusty.MakeKustomizer(fSys, opts)

	rm, err = k.Run(".")
	if err != nil {
		return nil, fmt.Errorf("Kustomizer Run failed: %s", err)
	}

	return rm, nil
}

func kustomizationOverlay(d *schema.ResourceData, m interface{}) error {

	var res []string
	for _, v := range d.Get("resources").([]interface{}) {
		res = append(res, v.(string))
	}

	kustomization := types.Kustomization{
		TypeMeta: types.TypeMeta{
			APIVersion: "kustomize.config.k8s.io/v1beta1",
			Kind:       "Kustomization",
		},
		Namespace: d.Get("namespace").(string),
		Resources: res,
	}

	rm, err := buildKustomizeOverlay(kustomization)
	if err != nil {
		return fmt.Errorf("buildKustomizeOverlay: %s", err)
	}

	d.Set("ids", flattenKustomizationIDs(rm))

	resources, err := flattenKustomizationResources(rm)
	if err != nil {
		return fmt.Errorf("buildKustomizeOverlay: %s", err)
	}
	d.Set("manifests", resources)

	id, err := getIDFromResources(rm)
	if err != nil {
		return fmt.Errorf("buildKustomizeOverlay: %s", err)
	}
	d.SetId(id)

	return nil
}
