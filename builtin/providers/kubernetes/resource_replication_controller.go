package kubernetes

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/yaml"
)

func resourceKubernetesReplicationController() *schema.Resource {

	s := resourceMeta()
	s["replicas"] = &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		//Required: true,
	}
	s["selector"] = &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		//Required: true,
	}

	s["pod"] = &schema.Schema{
		Type:     schema.TypeList, //this allows multiple values. should check for and reject that until I can figure something more clever
		Optional: true,
		Computed: true,
		ForceNew: true,
		Elem:     &schema.Resource{Schema: resourcePodSpec()},
	}

	//s["spec"] = &schema.Schema{
	//	Type:     schema.TypeString,
	//	//Required: true,
	//	Optional: true,
	//	Computed: true,
	//	StateFunc: func(input interface{}) string {
	//		src, err := normalizeReplicationControllerSpec(input.(string))
	//		if err != nil {
	//			log.Printf("[ERROR] Normalising spec failed: %q", err.Error())
	//		}
	//		return src
	//	},
	//}

	return &schema.Resource{
		Create: resourceKubernetesReplicationControllerCreate,
		Read:   resourceKubernetesReplicationControllerRead,
		Update: resourceKubernetesReplicationControllerUpdate,
		Delete: resourceKubernetesReplicationControllerDelete,

		Schema: s,
	}
}

func resourceKubernetesReplicationControllerCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)

	spec, err := constructReplicationControllerSpec(d)

	
	//spec, err := expandReplicationControllerSpec(d.Get("spec").(string))
	if err != nil {
		return err
	}

	//specString, err = flattenReplicationControllerSpec(spec)
	//if err != nil {
	//	return err
	//}
	//log.Printf("[DEBUG] rc spec: %v", specString)

	l := d.Get("labels").(map[string]interface{})
	labels := make(map[string]string, len(l))
	for k, v := range l {
		labels[k] = v.(string)
	}

	req := api.ReplicationController{
		ObjectMeta: api.ObjectMeta{
			Name:   d.Get("name").(string),
			Labels: labels,
		},
		Spec: spec,
	}

	ns := d.Get("namespace").(string)

	rc, err := c.ReplicationControllers(ns).Create(&req)
	if err != nil {
		return err
	}

	d.SetId(string(rc.UID))

	return nil // resourceKubernetesReplicationControllerRead(d, meta)
}

func resourceKubernetesReplicationControllerRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)
	rc, err := c.ReplicationControllers(d.Get("namespace").(string)).Get(d.Get("name").(string))
	if err != nil {
		return err
	}

	err = extractReplicationControllerSpec(d, rc)
	//spec, err := flattenReplicationControllerSpec(rc.Spec)
	if err != nil {
		return err
	}
	//d.Set("spec", spec)
	//d.Set("spec", rc.Spec)

	d.Set("labels", rc.Labels)
	d.Set("selector", rc.Spec.Selector)
	d.Set("replicas", rc.Spec.Replicas)

	return nil
}

func resourceKubernetesReplicationControllerUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)

	spec, err := expandReplicationControllerSpec(d.Get("spec").(string))
	if err != nil {
		return err
	}

	l := d.Get("labels").(map[string]interface{})
	labels := make(map[string]string, len(l))
	for k, v := range l {
		labels[k] = v.(string)
	}

	req := api.ReplicationController{
		ObjectMeta: api.ObjectMeta{
			Name:   d.Get("name").(string),
			Labels: labels,
		},
		Spec: spec,
	}

	_, err = c.ReplicationControllers(d.Get("namespace").(string)).Update(&req)
	if err != nil {
		return err
	}

	return resourceKubernetesReplicationControllerRead(d, meta)
}

func resourceKubernetesReplicationControllerDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)
	err := c.ReplicationControllers(d.Get("namespace").(string)).Delete(d.Get("name").(string))
	return err
}

func expandReplicationControllerSpec(input string) (spec api.ReplicationControllerSpec, err error) {
	r := strings.NewReader(input)
	y := yaml.NewYAMLOrJSONDecoder(r, 4096)

	err = y.Decode(&spec)
	if err != nil {
		return
	}
	return
}

func flattenReplicationControllerSpec(spec api.ReplicationControllerSpec) (string, error) {
	b, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func normalizeReplicationControllerSpec(input string) (string, error) {
	r := strings.NewReader(input)
	y := yaml.NewYAMLOrJSONDecoder(r, 4096)
	spec := api.ReplicationControllerSpec{}

	err := y.Decode(&spec)
	if err != nil {
		return "", err
	}

	// TODO: Add/ignore default structures, e.g. template.spec.restartPolicy = Always

	b, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func constructReplicationControllerSpec(d *schema.ResourceData) (spec api.ReplicationControllerSpec, err error) {

	template, err := constructPodRCSpec(d)
	if err != nil {
		var thing api.ReplicationControllerSpec
		return thing, err
	}
	
	var templateSpec api.PodTemplateSpec
	templateSpec.Spec = template

	label_map := make(map[string]string)
	for k, v := range d.Get("pod.0.labels").(map[string]interface{}) {
		log.Printf("[DEBUG]label: %#v %#v", k, v)
		label_map[k] = v.(string)
	}
	templateSpec.Labels = label_map

	spec.Template = &templateSpec

	spec.Replicas = d.Get("replicas").(int)

	selector_map := make(map[string]string)
	for k, v := range d.Get("selector").(map[string]interface{}) {
		log.Printf("[DEBUG]selector: %#v %#v", k, v)
		selector_map[k] = v.(string)
	}
	spec.Selector = selector_map

	return spec, err
}

func extractReplicationControllerSpec(d *schema.ResourceData, rc *api.ReplicationController) (err error) {
	d.Set("selector", rc.Spec.Selector)
	d.Set("labels", rc.ObjectMeta.Labels)

	pod, err := extractPodTemplateSpec(rc.Spec.Template) //I'm not abstracting this well enough. It's not portable to rc
	var pod_array [1]map[string]interface{}
	pod_array[0] = pod
	//var pod_array [1]string
	//pod_array[0] = "test"
	d.Set("pod", pod_array)

	return err
}
