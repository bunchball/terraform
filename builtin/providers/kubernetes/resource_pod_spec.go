package kubernetes

import (
//	"strconv"
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
)

func resourcePodSpec() map[string]*schema.Schema {
       return map[string]*schema.Schema{
  	     	"name": &schema.Schema{
  	     		Type:     schema.TypeString,
  	     		Required: true,
  	     		ForceNew: true,
  	     	},

  	     	"namespace": &schema.Schema{
  	     		Type:     schema.TypeString,
  	     		Optional: true,
  	     		ForceNew: true,
  	     		Default:  api.NamespaceDefault,
  	     	},

  	     	"labels": &schema.Schema{
  	     		Type:     schema.TypeMap,
  	     		Optional: true,
  	     	},

  	     	"nodeName": &schema.Schema{
  	     		Type:     schema.TypeString,
  	     		Optional: true,
  	     		Computed: true,
  	     	},

  	     	"terminationGracePeriodSeconds": &schema.Schema{
  	     		Type:     schema.TypeString,
  	     		Optional: true,
  	     		Computed: true,
  	     	},
  	     	"container": &schema.Schema{
  	     		Type:     schema.TypeList,
  	     		Optional: true,
  	     		Computed: true,
  	     		ForceNew: true,
  	     		Elem:     &schema.Resource{
  	     			Schema: map[string]*schema.Schema{
  	     				"name": &schema.Schema{
  	     					Type: schema.TypeString,
  	     					Required: true,
  	     					ForceNew: true,
  	     				},
  	     				"image": &schema.Schema{
  	     					Type: schema.TypeString,
  	     					Required: true,
  	     				},
  	     				"port": &schema.Schema{
  	     					Type: schema.TypeList,
  	     					Optional: true,
  	     					ForceNew: true,
  	     					Elem: &schema.Resource{
  	     						Schema: map[string]*schema.Schema{
  	     							"protocol": &schema.Schema{
  	     								Type: schema.TypeString,
  	     								Optional: true,
  	     								Default: "TCP",
  	     								ForceNew: true,
  	     							},
  	     							"containerPort": &schema.Schema{
  	     								Type: schema.TypeString,
  	     								Required: true,
  	     								ForceNew: true,
  	     							},
  	     							"name": &schema.Schema{
  	     								Type: schema.TypeString,
  	     								Required: true,
  	     								ForceNew: true,
  	     							},
  	     						},
  	     					},
  	     				},
  	     			},
  	     		},
  	     	},
       }

}

func constructPodSpec(d *schema.ResourceData) (spec api.PodSpec, err error) {
	containers := d.Get("container").([]interface{})
	for _, c_tf := range containers {
		c_tf_map := c_tf.(map[string]interface{})

		c, badSpec := constructContainerSpec(c_tf_map)
		if badSpec != nil {
			return
		}
		spec.Containers = append(spec.Containers, c)
	}
	
	return spec, err
}

func extractPodSpec(d *schema.ResourceData, pod *api.Pod) (err error) {
	d.Set("labels", pod.Labels)
	d.Set("nodeName", pod.Spec.NodeName)
	d.Set("terminationGracePeriodSeconds", pod.Spec.TerminationGracePeriodSeconds)

	var containers []map[string]interface{}
	for _, container := range pod.Spec.Containers {
		c, badContainer := extractContainerSpec(container)
		if badContainer != nil {
			return
		}
		containers = append(containers, c)
	}
	d.Set("container", containers)

	return nil
}
