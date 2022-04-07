package kustomize

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"sigs.k8s.io/kustomize/api/types"
)

func getHelmGlobals(hg map[string]interface{}) *types.HelmGlobals {
	g := &types.HelmGlobals{}
	g.ChartHome = hg["chart_home"].(string)
	g.ConfigHome = hg["config_home"].(string)

	return g
}

func getGeneratorOptionsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"annotations": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"disable_name_suffix_hash": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func getGeneratorOptions(o map[string]interface{}) *types.GeneratorOptions {
	g := &types.GeneratorOptions{}
	g.Labels = convertMapStringInterfaceToMapStringString(
		o["labels"].(map[string]interface{}),
	)

	g.Annotations = convertMapStringInterfaceToMapStringString(
		o["annotations"].(map[string]interface{}),
	)

	g.DisableNameSuffixHash = o["disable_name_suffix_hash"].(bool)

	return g
}

func getPatchOptionsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"allow_kind_change": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"allow_name_change": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func getPatchOptions(o map[string]interface{}) map[string]bool {
	po := make(map[string]bool)
	po["allowKindChange"] = o["allow_kind_change"].(bool)
	po["allowNameChange"] = o["allow_name_change"].(bool)
	return po
}

func dataSourceKustomizationOverlaySchemaV1() map[string]*schema.Schema {
	// support almost all attributes available in a Kustomization
	//
	// not implemented:
	// Bases (deprecated)
	// Configurations (incompatible with plugins)
	// Valid       ators (requires alpha plugins enabled, we only enable built-in plugins)
	return map[string]*schema.Schema{
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
						Optional: true,
					},
					"namespace": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"behavior": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice(
							[]string{"create", "replace", "merge"},
							false,
						),
					},
					"envs": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"files": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"literals": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"options": {
						Type:     schema.TypeList,
						MaxItems: 1,
						Optional: true,
						Elem:     getGeneratorOptionsSchema(),
					},
				},
			},
		},
		"crds": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"generators": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"generator_options": &schema.Schema{
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem:     getGeneratorOptionsSchema(),
		},
		"images": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"new_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"new_tag": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"digest": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"name_prefix": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"namespace": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"name_suffix": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"patches": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"options": {
						Type:     schema.TypeList,
						MaxItems: 1,
						Optional: true,
						Elem:     getPatchOptionsSchema(),
					},
					"path": {
						Type:     schema.TypeString,
						Optional: true,
						//ConflictsWith: []string{"patch"},
					},
					"patch": {
						Type:     schema.TypeString,
						Optional: true,
						//ConflictsWith: []string{"path"},
					},
					"target": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"group": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"version": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"kind": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"name": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"namespace": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"label_selector": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"annotation_selector": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
				},
			},
		},
		"replicas": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"count": {
						Type:     schema.TypeInt,
						Optional: true,
					},
				},
			},
		},
		"resources": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"helm_globals": &schema.Schema{
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"chart_home": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"config_home": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"helm_charts": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"version": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"repo": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"release_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"namespace": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"include_crds": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"values_merge": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice(
							[]string{"override", "replace", "merge"},
							false,
						),
					},
					"values_file": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"values_inline": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"secret_generator": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"namespace": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"behavior": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"type": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"envs": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"files": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"literals": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"options": {
						Type:     schema.TypeList,
						MaxItems: 1,
						Optional: true,
						Elem:     getGeneratorOptionsSchema(),
					},
				},
			},
		},
		"transformers": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"vars": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"obj_ref": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"api_version": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"group": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"version": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"kind": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"name": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"namespace": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
					"field_ref": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"field_path": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
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
		"kustomize_options": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"load_restrictor": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"enable_helm": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"helm_path": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}
}

func dataSourceKustomizationOverlayV0() *schema.Resource {
	return &schema.Resource{
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
							Optional: true,
						},
						"namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"behavior": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice(
								[]string{"create", "replace", "merge"},
								false,
							),
						},
						"envs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"files": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"literals": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"options": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem:     getGeneratorOptionsSchema(),
						},
					},
				},
			},
			"crds": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"generators": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"generator_options": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     getGeneratorOptionsSchema(),
			},
			"images": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"new_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"new_tag": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"digest": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"name_prefix": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"namespace": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"name_suffix": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"patches": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"options": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem:     getPatchOptionsSchema(),
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
							//ConflictsWith: []string{"patch"},
						},
						"patch": {
							Type:     schema.TypeString,
							Optional: true,
							//ConflictsWith: []string{"path"},
						},
						"target": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"version": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"kind": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"label_selector": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"annotation_selector": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"replicas": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"resources": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"helm_globals": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"chart_home": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"config_home": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"helm_charts": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"version": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"repo": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"release_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"include_crds": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"values_merge": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice(
								[]string{"override", "replace", "merge"},
								false,
							),
						},
						"values_file": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"values_inline": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"secret_generator": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"behavior": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"envs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"files": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"literals": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"options": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem:     getGeneratorOptionsSchema(),
						},
					},
				},
			},
			"transformers": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vars": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"obj_ref": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"api_version": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"group": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"version": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"kind": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"field_ref": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field_path": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
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
			"kustomize_options": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"load_restrictor": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enable_helm": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"helm_path": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceKustomizationOverlayStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	f, err := os.Create("_debug.log")
    if err != nil {
        panic(err)
    }

	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%v+", rawState))
	if err != nil {
        panic(err)
    }

	f.Sync()

	var emptyRawState map[string]interface{}

	return emptyRawState, nil
}
