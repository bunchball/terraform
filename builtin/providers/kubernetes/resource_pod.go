package kubernetes

import (
	"strings"
	"strconv"
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

func resourceKubernetesPod() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesPodCreate,
		Read:   resourceKubernetesPodRead,
		Update: resourceKubernetesPodUpdate,
		Delete: resourceKubernetesPodDelete,

		Schema: map[string]*schema.Schema{
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
		},
	}
}

func resourceKubernetesPodCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)

	spec, err := constructPodSpec(d)
	if err != nil {
		return err
	}

	l := d.Get("labels").(map[string]interface{})
	labels := make(map[string]string, len(l))
	for k, v := range l {
		labels[k] = v.(string)
	}

	req := api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:   d.Get("name").(string),
			Labels: labels,
		},
		Spec: spec,
	}

	ns := d.Get("namespace").(string)

	pod, err := c.Pods(ns).Create(&req)
	if err != nil {
		return err
	}

	d.SetId(string(pod.UID))

	return resourceKubernetesPodRead(d, meta)
}

func resourceKubernetesPodRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)
	pod, err := c.Pods(d.Get("namespace").(string)).Get(d.Get("name").(string))
	if err != nil {
		return err
	}

	d.Set("labels", pod.Labels)
	d.Set("nodeName", pod.Spec.NodeName)
	d.Set("terminationGracePeriodSeconds", pod.Spec.TerminationGracePeriodSeconds)

	var containers []map[string]interface{}
	for _, v := range pod.Spec.Containers {
		var container = make(map[string]interface{})
		container["name"] = v.Name
		container["image"] = v.Image
		var portList []interface{}
		for _, p := range v.Ports {
			var portMap = make(map[string]interface{})
			portMap["name"] = p.Name
			portMap["containerPort"] = strconv.Itoa(p.ContainerPort)
			portMap["protocol"] = p.Protocol 
			portList = append(portList, portMap)
		}
		container["port"] = portList
		containers = append(containers, container)
	}
	d.Set("container", containers)

	return nil
}

func resourceKubernetesPodUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)

	spec, err := constructPodSpec(d)
	if err != nil {
		return err
	}

	l := d.Get("labels").(map[string]interface{})
	labels := make(map[string]string, len(l))
	for k, v := range l {
		labels[k] = v.(string)
	}

	req := api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:   d.Get("name").(string),
			Labels: labels,
		},
		Spec: spec,
	}

	_, err = c.Pods(d.Get("namespace").(string)).Update(&req)
	if err != nil {
		return err
	}

	return resourceKubernetesPodRead(d, meta)
}

func resourceKubernetesPodDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)
	err := c.Pods(d.Get("namespace").(string)).Delete(d.Get("name").(string), nil)
	return err
}

func constructPodSpec(d *schema.ResourceData) (spec api.PodSpec, err error) {
	containers := d.Get("container").([]interface{})
	for _, c_tf := range containers {
		c_tf_map := c_tf.(map[string]interface{})

		var c api.Container
		c.Name = c_tf_map["name"].(string)
		c.Image = c_tf_map["image"].(string)

		ports := c_tf_map["port"].([]interface{})
		for _, p_tf := range ports {
			p_tf_map := p_tf.(map[string]interface{})

			var port api.ContainerPort
			port.Name = p_tf_map["name"].(string)

			portNumInt, notInt := strconv.Atoi(p_tf_map["containerPort"].(string))
			if notInt != nil {
				return
			}

			//not sure where to put error checking. will do later
			//if portNumInt > 1<<16 -1 {
			//	return
			//}

			port.ContainerPort = portNumInt

			switch protocol := strings.ToUpper(p_tf_map["protocol"].(string)); protocol {
				case "TCP":
					port.Protocol = api.ProtocolTCP
				case "UDP":
					port.Protocol = api.ProtocolUDP
				default:
					port.Protocol = api.ProtocolTCP
					//probably should error out here if something invalid is put
			}
			c.Ports = append(c.Ports, port)
		}
		spec.Containers = append(spec.Containers, c)
	}
	
	return spec, err
}
