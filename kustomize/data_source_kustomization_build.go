package kustomize

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"sigs.k8s.io/kustomize/api/filesys"
)

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

func kustomizationBuild(d *schema.ResourceData, m interface{}) error {
	path := d.Get("path").(string)

	fSys := filesys.MakeFsOnDisk()
	rm, err := runKustomizeBuild(fSys, path)
	if err != nil {
		return fmt.Errorf("kustomizationBuild: %s", err)
	}

	return setGeneratedAttributes(d, rm)
}
