package kustomize

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func dataSourceKustomizationOverlay() *schema.Resource {
	return &schema.Resource{
		Read: kustomizationOverlay,
       
		Schema: dataSourceKustomizationOverlaySchemaV1(),
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    dataSourceKustomizationOverlayV0().CoreConfigSchema().ImpliedType(),
				Upgrade: dataSourceKustomizationOverlayStateUpgradeV0,
				Version: 0,
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

	return setGeneratedAttributes(d, rm, m.(*Config).LegacyIDs)
}
