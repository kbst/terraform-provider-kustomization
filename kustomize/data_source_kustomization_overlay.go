package kustomize

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

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
				p["target"].(map[string]interface{}),
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

			k.Patches = append(k.Patches, kp)
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
				v["obj_ref"].(map[string]interface{}),
			)

			kv.ObjRef = types.Target{}
			kv.ObjRef.APIVersion = or["api_version"]
			kv.ObjRef.Group = or["group"]
			kv.ObjRef.Version = or["version"]
			kv.ObjRef.Kind = or["kind"]
			kv.ObjRef.Name = or["name"]
			kv.ObjRef.Namespace = or["namespace"]

			fr := convertMapStringInterfaceToMapStringString(
				v["field_ref"].(map[string]interface{}),
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
	rm, err := runKustomizeBuild(fSys, ".")
	mu.Unlock()
	if err != nil {
		return fmt.Errorf("buildKustomizeOverlay: %s", err)
	}

	return setGeneratedAttributes(d, rm)
}
