package kubernetes

import (
	"strconv"
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

	kPodTemplateSpec := rc.Spec.Template
	kPodSpec := kPodTemplateSpec.Spec
	var kc_holder []interface{}
	for _, cv := range kPodSpec.Containers {
		kc := make(map[string]interface{})
		kc["name"] = cv.Name
		kc["image"] = cv.Image
		var portList []interface{}
		for _, p := range cv.Ports {
			var portMap = make(map[string]interface{})
			portMap["name"] = p.Name
			portMap["containerPort"] = strconv.Itoa(p.ContainerPort)
			portMap["protocol"] = p.Protocol 
			portList = append(portList, portMap)
		}
		kc["port"] = portList
		kc_holder = append(kc_holder, kc)
	}
	kContainers := make(map[string]interface{})
	kContainers["container"] = kc_holder
	kContainers["labels"] = kPodTemplateSpec.Labels
	kContainers["nodeName"] = kPodSpec.NodeName
	kContainers["namespace"] = rc.Namespace
	kContainers["name"] = kPodTemplateSpec.Name //this currently doesn't work properly because pods are still inline and don't define a name
	kContainers["terminationGracePeriodSeconds"] = strconv.FormatInt(*kPodSpec.TerminationGracePeriodSeconds,10)
	var kPod []interface{}
	kPod = append(kPod, kContainers)
	d.Set("pod", kPod)

	return err
}
