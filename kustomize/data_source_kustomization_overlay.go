package kustomize

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func dataSourceKustomizationOverlay() *schema.Resource {
	return &schema.Resource{
		Read: kustomizationOverlay,

		Schema: map[string]*schema.Schema{
			"common_annotations": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"common_labels": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"components": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"config_map_generator": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"literals": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"namespace": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"resources": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
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

func convertListInterfaceToListString(in []interface{}) (out []string) {
	for _, v := range in {
		out = append(out, v.(string))
	}
	return out
}

func convertMapStringInterfaceToMapStringString(in map[string]interface{}) (out map[string]string) {
	out = make(map[string]string)
	for k, v := range in {
		out[k] = v.(string)
	}
	return out
}

func getKustomization(d *schema.ResourceData) (k types.Kustomization) {
	k.TypeMeta = types.TypeMeta{
		APIVersion: "kustomize.config.k8s.io/v1beta1",
		Kind:       "Kustomization",
	}

	if d.Get("common_annotations") != nil {
		k.CommonAnnotations = convertMapStringInterfaceToMapStringString(
			d.Get("common_annotations").(map[string]interface{}),
		)
	}

	if d.Get("common_labels") != nil {
		k.CommonLabels = convertMapStringInterfaceToMapStringString(
			d.Get("common_labels").(map[string]interface{}),
		)
	}

	if d.Get("components") != nil {
		k.Components = convertListInterfaceToListString(
			d.Get("components").([]interface{}),
		)
	}

	if d.Get("config_map_generator") != nil {
		cmgs := d.Get("config_map_generator").([]interface{})
		for i := range cmgs {
			cmg := cmgs[i].(map[string]interface{})
			cma := types.ConfigMapArgs{}
			cma.Name = cmg["name"].(string)
			cma.LiteralSources = convertListInterfaceToListString(
				cmg["literals"].([]interface{}),
			)
			k.ConfigMapGenerator = append(k.ConfigMapGenerator, cma)
		}
	}

	if d.Get("namespace") != nil {
		k.Namespace = d.Get("namespace").(string)
	}

	if d.Get("resources") != nil {
		k.Resources = convertListInterfaceToListString(
			d.Get("resources").([]interface{}),
		)
	}

	return k
}

func kustomizationOverlay(d *schema.ResourceData, m interface{}) error {
	k := getKustomization(d)

	fSys := filesys.MakeFsOnDisk()

	var b bytes.Buffer
	ye := yaml.NewEncoder(io.Writer(&b))
	ye.Encode(k)
	ye.Close()
	data, _ := ioutil.ReadAll(io.Reader(&b))

	fSys.WriteFile("Kustomization", data)
	defer fSys.RemoveAll("Kustomization")

	rm, err := runKustomizeBuild(fSys, ".")
	if err != nil {
		return fmt.Errorf("buildKustomizeOverlay: %s", err)
	}

	return setGeneratedAttributes(d, rm)
}
