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
			"kustomize_options": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"load_restrictor": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"ids": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      idSetHash,
			},
			"ids_prio": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				MinItems: 3,
				MaxItems: 3,
				Elem: &schema.Schema{
					Type: schema.TypeSet,
					Set:  idSetHash,
				},
			},
			"manifests": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
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

func kustomizationBuild(d *schema.ResourceData, m interface{}) error {
	path := d.Get("path").(string)

	kOpts := getKustomizeOptions(d)

	fSys := filesys.MakeFsOnDisk()

	// mutex as tmp workaround for upstream bug
	// https://github.com/kubernetes-sigs/kustomize/issues/3659
	mu := m.(*Config).Mutex
	mu.Lock()
	rm, err := runKustomizeBuild(fSys, path, kOpts.loadRestrictor)
	mu.Unlock()
	if err != nil {
		return fmt.Errorf("kustomizationBuild: %s", err)
	}

	return setGeneratedAttributes(d, rm)
}
