# Kustomize Provider

This provider allows building existing kustomizations using the `kustomization_build` data source or defining
dynamic kustomizations in HCL using the `kustomization_overlay` data source and applying the resources from
either kustomization against a Kubernetes cluster using the `kustomization_resource` resource.

The provider is maintained as part of the [Terraform GitOps framework Kubestack](https://www.kubestack.com/).

Using this Kustomize provider and Terraform has three main benefits compared to applying a kustomization using `kubectl`:

1. Running `terraform plan` will show a diff of the changes to be applied.
1. Deleted resources from a previous configuration will be purged.
1. Changes to immutable fields will generate a destroy and re-create plan.

As such the provider can be useful to replace kustomize/kubectl integrated into a Terraform configuration as a provisioner or to replace standalone `kubectl diff/apply` steps in CI/CD.

## Example Usage

```hcl
terraform {
  required_providers {
    kustomization = {
      source  = "kbst/kustomization"
      version = "0.9.0"
    }
  }
}

provider "kustomization" {
  # one of kubeconfig_path, kubeconfig_raw or kubeconfig_incluster must be set

  # kubeconfig_path = "~/.kube/config"
  # can also be set using KUBECONFIG_PATH environment variable

  # kubeconfig_raw = data.template_file.kubeconfig.rendered
  # kubeconfig_raw = yamlencode(local.kubeconfig)

  # kubeconfig_incluster = true
}

```

## Argument Reference

- `kubeconfig_path` - Path to a kubeconfig file. Can be set using `KUBECONFIG_PATH` environment variable.
- `kubeconfig_raw` - Raw kubeconfig file. If `kubeconfig_raw` is set, `kubeconfig_path` is ignored.
- `kubeconfig_incluster` - Set to `true` when running inside a kubernetes cluster.
- `context` - (Optional) Context to use in kubeconfig with multiple contexts, if not specified the default context is used.
- `legacy_id_format` - (Optional) Defaults to `false`. Provided for backward compability, set to `true` to use the legacy ID format. Removed starting `0.9.0`.
- `gzip_last_applied_config` - (Optional) Defaults to `true`. Use a gzip compressed and base64 encoded value for the lastAppliedConfig annotation if a resource would otherwise exceed the Kubernetes max annotation size. All other resources use the regular uncompressed annotation. Set to `false` to never use the compressed annotation.
- `exec` - (Optional) Configuration block to use an [exec-based credential plugin] (https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins), e.g. call an external command to receive user credentials.
    - `api_version` - (Required) API version to use when decoding the ExecCredentials resource, e.g. `client.authentication.k8s.io/v1beta1`.
    - `command` - (Required) Command to execute.
    - `args` - (Optional) List of arguments to pass when executing the plugin.
    - `env` - (Optional) Map of environment variables to set when executing the plugin.

## Migrating resource IDs from legacy format to format enabling API version upgrades

-> Support for the legacy ID format has been removed in version `0.9.0`. The provider has defaulted to the new format since version `0.7.0`. If you have been using the legacy format with the `legacy_id_format = true` backwards compatibility until now, make sure to migrate IDs before upgrading to `0.9.0`.

To allow the kustomization provider to manage API version upgrades, the version has been removed from resource IDs.
As this is a breaking change, we provide a helper script to move resources in the state.
If you choose not to move resources in the state, a destroy and recreate of all resources managed by the provider will be required.
We are also taking this as an opportunity, to refactor the ID format.

 * Legacy format: `apps_v1_Deployment|test-ns|test-deploy` or `~G_v1_Service|test-ns|test-svc`
 * New format: `apps/Deployment/test-ns/test-deploy` or `_/Service/test-ns/test-svc`

The general form is `group/Kind/namespace/name` with `_` as a placeholder for empty values (e.g. `_/Namespace/_/test-namespace`).

The commands below will create a file `state_mv.sh` with one `terraform state mv` command per resource.

```shell
cat > migrate.py <<'EOF'
import sys
import re

for si in sys.stdin.readlines():

    if not "kustomization_resource" in si:
        continue

    so = re.sub(r"(.*)\[\"([^_]*)\_[^_]*\_([^|]*)\|([^|]*)\|([^\"]*)\"\]",
                r'\1["\2/\3/\4/\5"]', si)

    so = re.sub(r"\~[GX]", r"_", so)

    print(f"terraform state mv '{si.strip()}' '{so.strip()}'")

EOF

terraform state list | python3 migrate.py > state_mv.sh
```

After carefully inspecting the generated commands in `state_mv.sh`, you can execute them using:

```shell
bash state_mv.sh
```

## Imports

To import existing resources, run `terraform import` as shown below.
Resources to be imported require a valid `kubectl.kubernetes.io/last-applied-configuration` annotation.

```
terraform import 'kustomization_resource.test["apps/Deployment/test-namespace/test-deployment"]' apps/Deployment/test-namespace/test-deployment
```
