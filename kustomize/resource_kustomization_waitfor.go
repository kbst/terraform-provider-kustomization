package kustomize

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smeta "k8s.io/apimachinery/pkg/api/meta"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

func waitForGVKCreated(d *schema.ResourceData, client dynamic.Interface, mapping *k8smeta.RESTMapping, namespace string, name string) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"existing"},
		Pending: []string{"pending"},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: func() (interface{}, string, error) {
			resp, err := client.
				Resource(mapping.Resource).
				Namespace(namespace).
				Get(context.TODO(), name, k8smetav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return nil, "pending", nil
				}
				return nil, "", err
			}

			return resp, "existing", nil
		},
	}

	return stateConf.WaitForState()
}
