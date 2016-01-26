package kubernetes

import (
	"fmt"
	"reflect"
	"log"
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
		//log.Fatal("nil container")
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
			log.Printf("[DEBUG] here2: %#v", c_tf_map["name"])
	
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

	//label_map := make(map[string]string)
	//for k, v := range d.Get("pod.0.label").(map[string]interface{}) {
	//	log.Printf("[DEBUG]label: %#v %#v", k, v)
	//	label_map[k] = v.(string)
	//}
	//spec.Labels = label_map

	nilTest := &spec
	if nilTest == nil {
		log.Printf("[DEBUG] here3: ")
		panic("nilTest!")
	}
	log.Printf("[DEBUG] here4: %#v", reflect.TypeOf(spec))

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

//func extractPodTemplateSpec(d *schema.ResourceData, pod *api.PodTemplateSpec) (err error) {
func extractPodTemplateSpec(pod *api.PodTemplateSpec) (pod_map map[string]interface{}, err error) {
	pod_map = make(map[string]interface{})
	//d.Set("labels", pod.Labels)
	pod_map["labels"] = pod.Labels
	//d.Set("nodeName", pod.Spec.NodeName)
	pod_map["nodeName"] = pod.Spec.NodeName

	//d.Set("terminationGracePeriodSeconds", pod.Spec.TerminationGracePeriodSeconds)
	pod_map["terminationGracePeriodSeconds"] = pod.Spec.TerminationGracePeriodSeconds

	var containers []map[string]interface{}
	for _, container := range pod.Spec.Containers {
		c, badContainer := extractContainerSpec(container)
		if badContainer != nil {
			return
		}
		containers = append(containers, c)
	}
	//d.Set("container", containers)
	pod_map["container"] = containers

	return pod_map, err
}

//func extractContainerSpec (v api.Container) (container map[string]interface{}, err error) {
//	container = make(map[string]interface{})
//	container["name"] = v.Name
//	container["image"] = v.Image
//	var portList []interface{}
//	for _, p := range v.Ports {
//		var portMap = make(map[string]interface{})
//		portMap["name"] = p.Name
//		portMap["containerPort"] = strconv.Itoa(p.ContainerPort)
//		portMap["protocol"] = p.Protocol 
//		portList = append(portList, portMap)
//	}
//	container["port"] = portList
//	err = nil
//	return container, err
//}
