package kustomize

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
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

func getReplacementSelectorSchema() *schema.Resource {
	return &schema.Resource{
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
		},
	}
}
func getReplacementOptionsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"delimiter": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"index": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"create": {
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

func dataSourceKustomizationOverlay() *schema.Resource {
	return &schema.Resource{
		Read: kustomizationOverlay,

		// support almost all attributes available in a Kustomization
		//
		// not implemented:
		// Bases (deprecated)
		// Configurations (incompatible with plugins)
		// Validators (requires alpha plugins enabled, we only enable built-in plugins)
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
			"labels": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pairs": {
							Type:     schema.TypeMap,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"include_selectors": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
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
			"replacements": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source": {
							Type:     schema.TypeList,
							MaxItems: 1,
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
									"field_path": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"options": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem:     getReplacementOptionsSchema(),
									},
								},
							},
						},
						"target": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"select": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem:     getReplacementSelectorSchema(),
									},
									"reject": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     getReplacementSelectorSchema(),
									},
									"field_paths": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"options": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem:     getReplacementOptionsSchema(),
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
						"skip_tests": {
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
						"api_versions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"kube_version": {
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
						"enable_alpha_plugins": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_exec": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_helm": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_star": {
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

func convertListInterfaceToListString(in []interface{}) (out []string) {
	for _, v := range in {
		out = append(out, v.(string))
	}
	return out
}

func convertListInterfaceFirstItemToMapStringInterface(in []interface{}) (out map[string]interface{}) {
	out = make(map[string]interface{})
	if len(in) > 0 {
		return in[0].(map[string]interface{})
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

func convertListInterfaceFirstItemToReplacementOptions(in []interface{}) (out *types.FieldOptions) {
	out = &types.FieldOptions{}
	if options := convertListInterfaceFirstItemToMapStringInterface(in); options != nil {
		if delimiter, ok := options["delimiter"]; ok {
			out.Delimiter = delimiter.(string)
		}
		if index, ok := options["index"]; ok {
			out.Index = index.(int)
		}
		if create, ok := options["create"]; ok {
			out.Create = create.(bool)
		}
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

	if d.Get("labels") != nil {
		lbls := d.Get("labels").([]interface{})
		for i := range lbls {
			if lbls[i] == nil {
				continue
			}
			lbl := lbls[i].(map[string]interface{})
			lb := types.Label{}

			lb.Pairs = convertMapStringInterfaceToMapStringString(
				lbl["pairs"].(map[string]interface{}),
			)
			lb.IncludeSelectors = lbl["include_selectors"].(bool)

			k.Labels = append(k.Labels, lb)
		}
	}

	if d.Get("components") != nil {
		k.Components = convertListInterfaceToListString(
			d.Get("components").([]interface{}),
		)
	}

	if d.Get("config_map_generator") != nil {
		cmgs := d.Get("config_map_generator").([]interface{})
		for i := range cmgs {
			if cmgs[i] == nil {
				continue
			}

			cmg := cmgs[i].(map[string]interface{})
			cma := types.ConfigMapArgs{}

			cma.Name = cmg["name"].(string)
			cma.Namespace = cmg["namespace"].(string)

			cma.Behavior = cmg["behavior"].(string)

			cma.EnvSources = convertListInterfaceToListString(
				cmg["envs"].([]interface{}),
			)

			cma.LiteralSources = convertListInterfaceToListString(
				cmg["literals"].([]interface{}),
			)

			cma.FileSources = convertListInterfaceToListString(
				cmg["files"].([]interface{}),
			)

			o := cmg["options"].([]interface{})
			if len(o) == 1 && o[0] != nil {
				cma.Options = getGeneratorOptions(o[0].(map[string]interface{}))
			}

			k.ConfigMapGenerator = append(k.ConfigMapGenerator, cma)
		}
	}

	if d.Get("transformers") != nil {
		k.Transformers = convertListInterfaceToListString(
			d.Get("transformers").([]interface{}),
		)
	}

	if d.Get("crds") != nil {
		k.Crds = convertListInterfaceToListString(
			d.Get("crds").([]interface{}),
		)
	}

	if d.Get("generators") != nil {
		k.Generators = convertListInterfaceToListString(
			d.Get("generators").([]interface{}),
		)
	}

	if d.Get("generator_options") != nil {
		gos := d.Get("generator_options").([]interface{})

		if len(gos) == 1 && gos[0] != nil {
			o := gos[0].(map[string]interface{})
			k.GeneratorOptions = getGeneratorOptions(o)
		}
	}

	if d.Get("images") != nil {
		imgs := d.Get("images").([]interface{})
		for i := range imgs {
			if imgs[i] == nil {
				continue
			}

			img := imgs[i].(map[string]interface{})
			kimg := types.Image{}

			kimg.Name = img["name"].(string)
			kimg.NewName = img["new_name"].(string)
			kimg.NewTag = img["new_tag"].(string)
			kimg.Digest = img["digest"].(string)

			k.Images = append(k.Images, kimg)
		}
	}

	if d.Get("patches") != nil {
		ps := d.Get("patches").([]interface{})
		for i := range ps {
			if ps[i] == nil {
				continue
			}

			p := ps[i].(map[string]interface{})
			kp := types.Patch{}

			kp.Path = p["path"].(string)
			kp.Patch = p["patch"].(string)

			t := convertMapStringInterfaceToMapStringString(
				convertListInterfaceFirstItemToMapStringInterface(
					p["target"].([]interface{}),
				),
			)

			if len(t) > 0 {
				kp.Target = &types.Selector{}
				kp.Target.Group = t["group"]
				kp.Target.Version = t["version"]
				kp.Target.Kind = t["kind"]
				kp.Target.Name = t["name"]
				kp.Target.Namespace = t["namespace"]
				kp.Target.AnnotationSelector = t["annotation_selector"]
				kp.Target.LabelSelector = t["label_selector"]
			}
			o := p["options"].([]interface{})
			if len(o) == 1 && o[0] != nil {
				kp.Options = getPatchOptions(o[0].(map[string]interface{}))
			}

			k.Patches = append(k.Patches, kp)
		}
	}

	if d.Get("replacements") != nil {
		rs := d.Get("replacements").([]interface{})
		for i := range rs {
			if rs[i] == nil {
				continue
			}

			r := rs[i].(map[string]interface{})
			kr := types.Replacement{}

			if path := r["path"].(string); path != "" {
				k.Replacements = append(k.Replacements, types.ReplacementField{Path: path})
				continue
			}
			source := convertListInterfaceFirstItemToMapStringInterface(r["source"].([]interface{}))

			kr.Source = &types.SourceSelector{
				ResId: resid.ResId{
					Gvk: resid.Gvk{
						Group:   source["group"].(string),
						Version: source["version"].(string),
						Kind:    source["kind"].(string),
					},
					Name:      source["name"].(string),
					Namespace: source["namespace"].(string),
				},
				FieldPath: source["field_path"].(string),
				Options:   convertListInterfaceFirstItemToReplacementOptions(source["options"].([]interface{})),
			}
			targets := r["target"].([]interface{})
			kr.Targets = make([]*types.TargetSelector, len(targets))

			for i, tgt := range targets {
				target := tgt.(map[string]interface{})
				kr.Targets[i] = &types.TargetSelector{
					Options:    convertListInterfaceFirstItemToReplacementOptions(target["options"].([]interface{})),
					FieldPaths: convertListInterfaceToListString(target["field_paths"].([]interface{})),
				}
				selector := convertMapStringInterfaceToMapStringString(convertListInterfaceFirstItemToMapStringInterface(target["select"].([]interface{})))
				kr.Targets[i].Select = &types.Selector{
					ResId: resid.ResId{
						Gvk: resid.Gvk{
							Group:   selector["group"],
							Version: selector["version"],
							Kind:    selector["kind"],
						},
						Name:      selector["name"],
						Namespace: selector["namespace"],
					},
				}
				rejects := target["reject"].([]interface{})
				kr.Targets[i].Reject = make([]*types.Selector, len(rejects))
				for j, rj := range rejects {
					reject := convertMapStringInterfaceToMapStringString((rj.(map[string]interface{})))
					kr.Targets[i].Reject[j] = &types.Selector{
						ResId: resid.ResId{
							Gvk: resid.Gvk{
								Group:   reject["group"],
								Version: reject["version"],
								Kind:    reject["kind"],
							},
							Name:      reject["name"],
							Namespace: reject["namespace"],
						},
					}
				}

			}
			k.Replacements = append(k.Replacements, types.ReplacementField{Replacement: kr})
		}
	}

	if d.Get("replicas") != nil {
		rs := d.Get("replicas").([]interface{})
		for i := range rs {
			if rs[i] == nil {
				continue
			}

			img := rs[i].(map[string]interface{})
			r := types.Replica{}

			r.Name = img["name"].(string)
			r.Count = int64(img["count"].(int))

			k.Replicas = append(k.Replicas, r)
		}
	}

	if d.Get("name_prefix") != nil {
		k.NamePrefix = d.Get("name_prefix").(string)
	}

	if d.Get("namespace") != nil {
		k.Namespace = d.Get("namespace").(string)
	}

	if d.Get("name_suffix") != nil {
		k.NameSuffix = d.Get("name_suffix").(string)
	}

	if d.Get("resources") != nil {
		k.Resources = convertListInterfaceToListString(
			d.Get("resources").([]interface{}),
		)
	}

	if d.Get("helm_globals") != nil {
		hgs := d.Get("helm_globals").([]interface{})

		if len(hgs) == 1 && hgs[0] != nil {
			hg := hgs[0].(map[string]interface{})
			k.HelmGlobals = getHelmGlobals(hg)
		}
	}

	if d.Get("helm_charts") != nil {
		hcs := d.Get("helm_charts").([]interface{})
		for i := range hcs {
			if hcs[i] == nil {
				continue
			}

			hc := hcs[i].(map[string]interface{})
			hca := types.HelmChart{}

			hca.Name = hc["name"].(string)
			hca.Version = hc["version"].(string)
			hca.Repo = hc["repo"].(string)
			hca.ReleaseName = hc["release_name"].(string)
			hca.Namespace = hc["namespace"].(string)
			hca.ValuesFile = hc["values_file"].(string)
			hca.ValuesMerge = hc["values_merge"].(string)
			hca.IncludeCRDs = hc["include_crds"].(bool)
			hca.SkipTests = hc["skip_tests"].(bool)
			hca.KubeVersion = hc["kube_version"].(string)
			hca.ApiVersions = convertListInterfaceToListString(hc["api_versions"].([]interface{}))

			hc_vi := make(map[string]interface{})
			if err := yaml.Unmarshal([]byte(hc["values_inline"].(string)), &hc_vi); err != nil {
				fmt.Printf("error: %v", err)
			}
			hca.ValuesInline = hc_vi

			k.HelmCharts = append(k.HelmCharts, hca)
		}
	}

	if d.Get("secret_generator") != nil {
		sg := d.Get("secret_generator").([]interface{})
		for i := range sg {
			if sg[i] == nil {
				continue
			}

			s := sg[i].(map[string]interface{})
			sa := types.SecretArgs{}

			sa.Name = s["name"].(string)
			sa.Namespace = s["namespace"].(string)

			sa.Behavior = s["behavior"].(string)
			sa.Type = s["type"].(string)

			sa.EnvSources = convertListInterfaceToListString(
				s["envs"].([]interface{}),
			)

			sa.LiteralSources = convertListInterfaceToListString(
				s["literals"].([]interface{}),
			)

			sa.FileSources = convertListInterfaceToListString(
				s["files"].([]interface{}),
			)

			o := s["options"].([]interface{})
			if len(o) == 1 && o[0] != nil {
				sa.Options = getGeneratorOptions(o[0].(map[string]interface{}))
			}

			k.SecretGenerator = append(k.SecretGenerator, sa)
		}
	}

	if d.Get("vars") != nil {
		vs := d.Get("vars").([]interface{})
		for i := range vs {
			if vs[i] == nil {
				continue
			}

			v := vs[i].(map[string]interface{})
			kv := types.Var{}

			kv.Name = v["name"].(string)

			or := convertMapStringInterfaceToMapStringString(
				convertListInterfaceFirstItemToMapStringInterface(
					v["obj_ref"].([]interface{}),
				),
			)

			kv.ObjRef = types.Target{}
			kv.ObjRef.APIVersion = or["api_version"]
			kv.ObjRef.Group = or["group"]
			kv.ObjRef.Version = or["version"]
			kv.ObjRef.Kind = or["kind"]
			kv.ObjRef.Name = or["name"]
			kv.ObjRef.Namespace = or["namespace"]

			fr := convertMapStringInterfaceToMapStringString(
				convertListInterfaceFirstItemToMapStringInterface(
					v["field_ref"].([]interface{}),
				),
			)

			kv.FieldRef = types.FieldSelector{}
			kv.FieldRef.FieldPath = fr["field_path"]

			k.Vars = append(k.Vars, kv)
		}
	}

	return k
}

func refuseExistingKustomization(fSys filesys.FileSystem) error {
	for _, n := range konfig.RecognizedKustomizationFileNames() {
		if fSys.Exists(n) {
			return fmt.Errorf("buildKustomizeOverlay: Can not build dynamic overlay, found %q in working directory.", n)
		}
	}
	return nil
}

func kustomizationOverlay(d *schema.ResourceData, m interface{}) error {
	k := getKustomization(d)

	var b bytes.Buffer
	ye := yaml.NewEncoder(io.Writer(&b))
	ye.Encode(k)
	ye.Close()
	data, _ := ioutil.ReadAll(io.Reader(&b))

	fSys, tmp, err := makeOverlayFS(filesys.MakeFsOnDisk())
	defer os.RemoveAll(tmp)
	if err != nil {
		return err
	}

	// error if the current working directory is already a Kustomization
	err = refuseExistingKustomization(fSys)
	if err != nil {
		return err
	}

	fSys.WriteFile(KFILENAME, data)
	defer fSys.RemoveAll(KFILENAME)

	// mutex as tmp workaround for upstream bug
	// https://github.com/kubernetes-sigs/kustomize/issues/3659
	mu := m.(*Config).Mutex
	mu.Lock()
	rm, err := runKustomizeBuild(fSys, ".", d)
	mu.Unlock()
	if err != nil {
		return fmt.Errorf("buildKustomizeOverlay: %s", err)
	}

	return setGeneratedAttributes(d, rm)
}
