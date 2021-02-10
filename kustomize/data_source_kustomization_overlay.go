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

func getKustomization(d *schema.ResourceData) (k types.Kustomization, err error) {

	var res []string
	rdRes := d.Get("resources")
	if rdRes != nil {
		for _, v := range rdRes.([]interface{}) {
			res = append(res, v.(string))
		}
	}

	cas := make(map[string]string)
	rdCA := d.Get("common_annotations")
	if rdCA != nil {
		for k, v := range rdCA.(map[string]interface{}) {
			cas[k] = v.(string)
		}
	}

	cls := make(map[string]string)
	rdCL := d.Get("common_labels")
	if rdCL != nil {
		for k, v := range rdCL.(map[string]interface{}) {
			cls[k] = v.(string)
		}
	}

	k = types.Kustomization{
		TypeMeta: types.TypeMeta{
			APIVersion: "kustomize.config.k8s.io/v1beta1",
			Kind:       "Kustomization",
		},
		CommonAnnotations: cas,
		CommonLabels:      cls,
		Namespace:         d.Get("namespace").(string),
		Resources:         res,
	}

	return k, nil
}

func kustomizationOverlay(d *schema.ResourceData, m interface{}) error {
	k, _ := getKustomization(d)

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
