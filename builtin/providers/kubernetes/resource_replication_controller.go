package kubernetes

import (
	"log"
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

func resourceKubernetesReplicationController() *schema.Resource {

	s := resourceMeta()
	s["replicas"] = &schema.Schema{
		Type:     schema.TypeInt,
		Required: true,
	}
	s["selector"] = &schema.Schema{
		Type:     schema.TypeMap,
		Required: true,
	}

	s["pod"] = &schema.Schema{
		Type:     schema.TypeList, //this allows multiple values. should check for and reject that until I can figure something more clever
		Optional: true,
		Computed: true,
		ForceNew: true,
		Elem:     &schema.Resource{Schema: resourcePodSpec()},
	}

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

	ns := d.Get("namespace").(string)

	rc, err := c.ReplicationControllers(ns).Create(&req)
	if err != nil {
		return err
	}

	d.SetId(string(rc.UID))

	return resourceKubernetesReplicationControllerRead(d, meta)
}

func resourceKubernetesReplicationControllerRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)
	rc, err := c.ReplicationControllers(d.Get("namespace").(string)).Get(d.Get("name").(string))
	if err != nil {
		return err
	}

	err = extractReplicationControllerSpec(d, rc)
	if err != nil {
		return err
	}

	return nil
}

func resourceKubernetesReplicationControllerUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)

	spec, err := constructReplicationControllerSpec(d)
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

	kContainers, c_err := extractPodTemplateSpec(d, rc.Spec.Template)
	if c_err != nil {
		return c_err
	}
	
	kContainers["namespace"] = rc.Namespace
	var kPod []interface{}
	kPod = append(kPod, kContainers)
	d.Set("pod", kPod)

	return err
}
