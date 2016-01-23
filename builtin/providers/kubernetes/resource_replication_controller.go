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

	s["spec"] = &schema.Schema{
		Type:     schema.TypeString,
		//Required: true,
		Optional: true,
		Computed: true,
		StateFunc: func(input interface{}) string {
			src, err := normalizeReplicationControllerSpec(input.(string))
			if err != nil {
				log.Printf("[ERROR] Normalising spec failed: %q", err.Error())
			}
			return src
		},
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

	//spec, err := constructReplicationControllerSpec(d)
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
	template, err := constructPodSpec(d)
	spec.Template.Spec = template
	spec.Replicas = d.Get("replicas").(int)
	spec.Selector = d.Get("selector").(map[string]string)

	return spec, err
}

func extractReplicationControllerSpec(d *schema.ResourceData, rc *api.ReplicationController) (err error) {
	d.Set("selector", rc.Spec.Selector)
	d.Set("labels", rc.ObjectMeta.Labels)
	d.Set("pod", rc.Spec.Template.Spec) //ObjectMeta isn't handled properly...

	//var containers []map[string]interface{}
	//for _, container := range pod.Spec.Containers {
	//	c, badContainer := extractContainerSpec(container)
	//	if badContainer != nil {
	//		return
	//	}
	//	containers = append(containers, c)
	//}
	//d.Set("container", containers)

	return nil
}
