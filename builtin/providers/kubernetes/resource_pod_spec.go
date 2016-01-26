package kubernetes

import (
	"strconv"
	"fmt"
	"reflect"
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
)

func resourcePodSpec() map[string]*schema.Schema {
	s := resourceMeta()

	s["nodeName"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	}

	s["terminationGracePeriodSeconds"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	}

	s["container"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		ForceNew: true,
		Elem:     &schema.Resource{Schema: resourceContainerSpec()},
	}
	return s
}

func constructPodSpec(d *schema.ResourceData) (spec api.PodSpec, err error) {
	
	containers := d.Get("container")
	if containers != nil {
		containers := d.Get("container").([]interface{})

		for _, c_tf := range containers {
			c_tf_map := c_tf.(map[string]interface{})
	
			c, badSpec := constructContainerSpec(c_tf_map)
			if badSpec != nil {
				return
			}
			spec.Containers = append(spec.Containers, c)
		}
	} else {
		panic("nil container")
	}

	return spec, err
}

func constructPodRCSpec(d *schema.ResourceData) (spec api.PodSpec, err error) {
	
	containers := d.Get("pod.0.container")
	if containers != nil {
		containers := d.Get("pod.0.container").([]interface{})

		fmt.Println("here")
		fmt.Println(reflect.TypeOf(containers))
		for _, c_tf := range containers {
			c_tf_map := c_tf.(map[string]interface{})
			//log.Printf("[DEBUG] here2: %#v", c_tf_map["name"])
	
			c, badSpec := constructContainerSpec(c_tf_map)
			if badSpec != nil {
				return
			}
			spec.Containers = append(spec.Containers, c)
		}
	} else {
		//log.Fatal("nil container")
		panic("nil container")
	}

	nilTest := &spec
	if nilTest == nil {
		panic("nilTest!")
	}

	return spec, err
}

//this almost certainly can be combined with the extractPodTemplateSpec somehow
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

func extractPodTemplateSpec(d *schema.ResourceData, pod *api.PodTemplateSpec) (pod_map map[string]interface{}, err error) {

	var c_holder []interface{}
	for _, cv := range pod.Spec.Containers {
		kc, c_err := extractContainerSpec (cv)
		if c_err != nil {
			var nilMap map[string]interface{}
			return nilMap, c_err
		}
		c_holder = append(c_holder, kc)
	}

	pod_map = make(map[string]interface{})
	pod_map["container"] = c_holder
	pod_map["labels"] = pod.Labels
	pod_map["nodeName"] = pod.Spec.NodeName
	pod_map["namespace"] = pod.Namespace //currently not set because pods are inline to RC. handled on the RC side
	pod_map["name"] = pod.Name //this currently doesn't work properly because pods are still inline and don't define a name
	pod_map["terminationGracePeriodSeconds"] = strconv.FormatInt(*pod.Spec.TerminationGracePeriodSeconds,10)

	return pod_map, err
}
