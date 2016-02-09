package kubernetes

import (
	"strings"
	"reflect"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/yaml"
	"k8s.io/kubernetes/pkg/util"
)

func resourceKubernetesService() *schema.Resource {
	s := resourceMeta()

	s["selector"] = &schema.Schema{
		Type:     schema.TypeMap,
		Required: true,
	}

	s["clusterIP"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	}

	s["type"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		ForceNew: true,
	}

	s["port"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"protocol": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "TCP",
					ForceNew: true,
				},
				"port": &schema.Schema{
					Type:     schema.TypeInt,
					Required: true,
					ForceNew: true,
				},
				"targetPort": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"nodePort": &schema.Schema{
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},
		},
	}

	return &schema.Resource{
		Create: resourceKubernetesServiceCreate,
		Read:   resourceKubernetesServiceRead,
		Update: resourceKubernetesServiceUpdate,
		Delete: resourceKubernetesServiceDelete,
		Schema: s,
	}
}

func resourceKubernetesServiceCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)

	spec, err := constructServiceSpec(d)
	if err != nil {
		return err
	}

	l := d.Get("labels").(map[string]interface{})
	labels := make(map[string]string, len(l))
	for k, v := range l {
		labels[k] = v.(string)
	}

	req := api.Service{
		ObjectMeta: api.ObjectMeta{
			Name:   d.Get("name").(string),
			Labels: labels,
		},
		Spec: spec,
	}

	ns := d.Get("namespace").(string)

	svc, err := c.Services(ns).Create(&req)
	if err != nil {
		return err
	}

	d.SetId(string(svc.UID))

	return resourceKubernetesServiceRead(d, meta)
}

func resourceKubernetesServiceRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)
	svc, err := c.Services(d.Get("namespace").(string)).Get(d.Get("name").(string))
	if err != nil {
		return err
	}

	d.Set("labels", svc.Labels)
	d.Set("selector", svc.Spec.Selector)
	d.Set("clusterIP", svc.Spec.ClusterIP)
	d.Set("type", svc.Spec.Type)

	switch stype := svc.Spec.Type; stype {
		case "NodePort":
			d.Set("type", "NodePort")
		case "ClusterIP":
			d.Set("type", "ClusterIP")
		case "LoadBalancer":
			d.Set("type", "LoadBalancer")
		default:
			log.Printf("[ERROR] Unknown Kubernetes Service Type: %q", err.Error())
	}

	var ports []map[string]interface{}
	for _, v := range svc.Spec.Ports {
		port := make(map[string]interface{})
		port["name"] = v.Name
		port["port"] = v.Port
		if &v.NodePort != nil {
			port["nodePort"] = v.NodePort
		}
		port["protocol"] = v.Protocol
		port["targetPort"] = v.TargetPort.String()
		ports = append(ports, port)
	}
	d.Set("port", ports)

	return nil
}

func resourceKubernetesServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)

	spec, err := expandServiceSpec(d.Get("spec").(string))
	if err != nil {
		return err
	}

	l := d.Get("labels").(map[string]interface{})
	labels := make(map[string]string, len(l))
	for k, v := range l {
		labels[k] = v.(string)
	}

	req := api.Service{
		ObjectMeta: api.ObjectMeta{
			Name:   d.Get("name").(string),
			Labels: labels,
		},
		Spec: spec,
	}

	_, err = c.Services(d.Get("namespace").(string)).Update(&req)
	if err != nil {
		return err
	}

	return resourceKubernetesServiceRead(d, meta)
}

func resourceKubernetesServiceDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)
	err := c.Services(d.Get("namespace").(string)).Delete(d.Get("name").(string))
	return err
}

func expandServiceSpec(input string) (spec api.ServiceSpec, err error) {
	r := strings.NewReader(input)
	y := yaml.NewYAMLOrJSONDecoder(r, 4096)

	err = y.Decode(&spec)
	if err != nil {
		return
	}

	return
}

func constructServiceSpec(d *schema.ResourceData) (spec api.ServiceSpec, err error) {
	
	selector := make(map[string]string)
	for k, s := range d.Get("selector").(map[string]interface{}) {
		selector[k] = s.(string)
	}
	spec.Selector = selector
	spec.ClusterIP = d.Get("clusterIP").(string)

	switch stype := d.Get("type").(string); stype {
		case "NodePort":
			spec.Type = api.ServiceTypeNodePort
		case "ClusterIP":
			spec.Type = api.ServiceTypeClusterIP
		case "LoadBalancer":
			spec.Type = api.ServiceTypeLoadBalancer
		case "":
			//nothing to do. kubernetes will use the default
		default:
			log.Printf("[DEBUG] Unknown Kubernetes Service Type: %s", stype)
	}
			

	var ports []api.ServicePort
	for _, p := range d.Get("port").([]interface{}) {
		p_map := p.(map[string]interface{})

		var port api.ServicePort
		port.Name = p_map["name"].(string)

		switch protocol := strings.ToUpper(p_map["protocol"].(string)); protocol {
			case "TCP":
				port.Protocol = api.ProtocolTCP
			case "UDP":
				port.Protocol = api.ProtocolUDP
			default:
				port.Protocol = api.ProtocolTCP
				//probably should error out here if something invalid is put
		}

		portNumInt := p_map["port"].(int)
		port.Port = portNumInt

		//this isn't going to work like you think. It's always a string, so if you put an int, kube will get mad
		switch typ := reflect.TypeOf(p_map["targetPort"]).Kind(); typ {
			case reflect.String:
				port.TargetPort = util.NewIntOrStringFromString(p_map["targetPort"].(string))
			case reflect.Int:
				port.TargetPort = util.NewIntOrStringFromInt(p_map["targetPort"].(int))
			default:
				panic("Not a string or int")	
		}

		if nodePort, ok := p_map["nodePort"]; ok {
			port.NodePort = nodePort.(int)
		}

		ports = append(ports, port)
	}
	spec.Ports = ports
	
	return spec, err
}
