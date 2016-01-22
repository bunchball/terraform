package kubernetes

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/yaml"

	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/hashcode"

	"strconv"
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
			"containers": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type: schema.TypeString,
							Required: true,
						},
						"image": &schema.Schema{
							Type: schema.TypeString,
							Required: true,
						},
						"ports": &schema.Schema{
							Type: schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"protocol": &schema.Schema{
										Type: schema.TypeString,
										Optional: true,
										Default: "tcp",
									},
									"containerPort": &schema.Schema{
										Type: schema.TypeString,
										Required: true,
									},
									"name": &schema.Schema{
										Type: schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"spec": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(input interface{}) string {
					s, err := normalizePodSpec(input.(string))
					if err != nil {
						log.Printf("[ERROR] Normalising spec failed: %q", err.Error())
					}
					return s
				},
			},
		},
	}
}

func resourceKubernetesPodCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)

	spec, err := expandPodSpec(d.Get("spec").(string))
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

type containerStruct struct {
	Name string
	Image string
	Ports map[string]string
}

func resourceKubernetesPodRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)
	pod, err := c.Pods(d.Get("namespace").(string)).Get(d.Get("name").(string))
	if err != nil {
		return err
	}

	//spec, err := flattenPodSpec(pod.Spec)
	//if err != nil {
	//	return err
	//}
	//d.Set("spec", spec)
	d.Set("labels", pod.Labels)
	//d.Set("spec", pod.Spec)
	d.Set("nodeName", pod.Spec.NodeName)
	d.Set("terminationGracePeriodSeconds", pod.Spec.TerminationGracePeriodSeconds)

	var containers []interface{}
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
		container["ports"] = portList
		containers = append(containers, container)
	}
	d.Set("containers", containers)

	return nil
}

func resourceKubernetesPodUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*client.Client)

	spec, err := expandPodSpec(d.Get("spec").(string))
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

func expandPodSpec(input string) (spec api.PodSpec, err error) {
	r := strings.NewReader(input)
	y := yaml.NewYAMLOrJSONDecoder(r, 4096)

	err = y.Decode(&spec)
	if err != nil {
		return
	}
	spec = setDefaultPodSpecValues(&spec)
	return
}

func flattenPodSpec(spec api.PodSpec) (string, error) {
	b, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func normalizePodSpec(input string) (string, error) {
	r := strings.NewReader(input)
	y := yaml.NewYAMLOrJSONDecoder(r, 4096)
	spec := api.PodSpec{}

	err := y.Decode(&spec)
	if err != nil {
		return "", err
	}

	spec = setDefaultPodSpecValues(&spec)

	b, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// This is to prevent detecting change when there's nothing to change
func setDefaultPodSpecValues(spec *api.PodSpec) api.PodSpec {
	if spec.ServiceAccountName == "" {
		spec.ServiceAccountName = "default"
	}
	if spec.RestartPolicy == "" {
		spec.RestartPolicy = "Always"
	}
	if spec.DNSPolicy == "" {
		spec.DNSPolicy = "ClusterFirst"
	}

	for k, c := range spec.Containers {
		if c.ImagePullPolicy == "" {
			spec.Containers[k].ImagePullPolicy = "IfNotPresent"
		}
		if c.TerminationMessagePath == "" {
			spec.Containers[k].TerminationMessagePath = "/dev/termination-log"
		}

		for _, p := range c.Ports {
			if p.Protocol == "" {
				p.Protocol = "TCP"
			}
		}
	}

	return *spec
}

func resourceContainerHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	//setting the hash based on the name only, insuffient but works for now
	if v,ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%v-", v.(string)))
	}

	return hashcode.String(buf.String())
}
