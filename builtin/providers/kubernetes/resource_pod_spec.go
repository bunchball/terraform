package kubernetes

import (
	"log"
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

	s["volume"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				//here follows the types of volumeSources. Parameter checking is mostly handled by the client for now
				//every type is multually exclusive with every other type
				"emptyDir": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					ForceNew: true,
					//Default:  make(map[string]interface{}),
					//ConflictsWith: []string{
					//	"hostPath",
					//},
				},
				"hostPath": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					ForceNew: true,
					//Default:  make(map[string]interface{}),
					//ConflictsWith: []string{
					//	"emptyDir",
					//},
				},
				"awsElasticBlockStore": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					ForceNew: true,
					//Default:  make(map[string]interface{}),
					//ConflictsWith: []string{
					//	"emptyDir",
					//},
				},
			},
		},
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

	volumes := d.Get("volume").([]interface{})
	for k, v := range volumes {
		log.Printf("[DEBUG] here k: %#v v %#v", k, v)
	}
		
	for vk, v := range volumes {
		v_map := v.(map[string]interface{})
		var vol api.Volume
		vol.Name = v_map["name"].(string)
		var vs api.VolumeSource
		log.Printf("[DEBUG] volume: %#v", vk)
		switch {
			case len(v_map["emptyDir"].(map[string]interface{})) > 0:
				var vi api.EmptyDirVolumeSource
				switch m := v_map["emptyDir"].(map[string]interface{})["medium"].(string); m {
					case "Memory":
						vi.Medium = api.StorageMediumMemory
					default:
						vi.Medium = api.StorageMediumDefault
				}
				vs.EmptyDir = &vi
			case len(v_map["hostPath"].(map[string]interface{})) > 0:
				var vi api.HostPathVolumeSource
				for tk, tv := range v_map["hostPath"].(map[string]interface{}) {
					log.Printf("[DEBUG] k: %#v v: %#v", tk, tv)
				}
				switch v_map["hostPath"].(map[string]interface{})["path"] {
					case nil:
						log.Printf("[DEBUG] hostPath: %#v", v_map["hostPath"])
						panic("path shouldn't be nil")
					default:
						vi.Path = v_map["hostPath"].(map[string]interface{})["path"].(string)
				}
				vs.HostPath = &vi
			case len(v_map["awsElasticBlockStore"].(map[string]interface{})) > 0:
				var vi api.AWSElasticBlockStoreVolumeSource
				awsElasticBlockStore_map := v_map["awsElasticBlockStore"].(map[string]interface{})
				switch awsElasticBlockStore_map["volumeID"] {
					case nil:
						panic("volumeID shouldn't be nil")
					default:
						vi.VolumeID  = awsElasticBlockStore_map["volumeID"].(string)
				}
				switch awsElasticBlockStore_map["fsType"] {
					case nil:
					default:
						vi.FSType = awsElasticBlockStore_map["fsType"].(string)
				}
				switch awsElasticBlockStore_map["partition"] {
					case nil:
					default:
						vi.Partition = awsElasticBlockStore_map["partition"].(int)
				}
				switch awsElasticBlockStore_map["readOnly"] {
					case nil:
					default:
						vi.ReadOnly = awsElasticBlockStore_map["readOnly"].(bool)
				}
				vs.AWSElasticBlockStore = &vi
			default:
				panic("not a known volumeSource type")
		}
		vol.VolumeSource = vs
		
		spec.Volumes = append(spec.Volumes, vol)
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

	volumes := d.Get("pod.0.volume").([]interface{})
	for _, v := range volumes {
		v_map := v.(map[string]interface{})
		var vol api.Volume
		vol.Name = v_map["name"].(string)
		var vs api.VolumeSource
		switch {
			//case v_map["emptyDir"] != nil:
			case len(v_map["emptyDir"].(map[string]interface{})) > 0:
				var vi api.EmptyDirVolumeSource
				switch m := v_map["emptyDir"].(map[string]interface{})["medium"]; m {
					case "Memory":
						vi.Medium = api.StorageMediumMemory
					default:
						vi.Medium = api.StorageMediumDefault
				}
				vs.EmptyDir = &vi
			//case v_map["hostPath"] != nil:
			case len(v_map["hostPath"].(map[string]interface{})) > 0:
				var vi api.HostPathVolumeSource
				vi.Path = v_map["hostPath"].(map[string]interface{})["path"].(string)
				vs.HostPath = &vi
			case len(v_map["awsElasticBlockStore"].(map[string]interface{})) > 0:
				var vi api.AWSElasticBlockStoreVolumeSource
				awsElasticBlockStore_map := v_map["awsElasticBlockStore"].(map[string]interface{})
				switch awsElasticBlockStore_map["volumeID"] {
					case nil:
						panic("volumeID shouldn't be nil")
					default:
						vi.VolumeID  = awsElasticBlockStore_map["volumeID"].(string)
				}
				switch awsElasticBlockStore_map["fsType"] {
					case nil:
					default:
						vi.FSType = awsElasticBlockStore_map["fsType"].(string)
				}
				switch awsElasticBlockStore_map["partition"] {
					case nil:
					default:
						vi.Partition = awsElasticBlockStore_map["partition"].(int)
				}
				switch awsElasticBlockStore_map["readOnly"] {
					case nil:
					default:
						vi.ReadOnly = awsElasticBlockStore_map["readOnly"].(bool)
				}
				vs.AWSElasticBlockStore = &vi
			default:
				panic("not a known volumeSource type")
		}
		vol.VolumeSource = vs
		
		spec.Volumes = append(spec.Volumes, vol)
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

	var volumes []map[string]interface{}
	for _, volume := range pod.Spec.Volumes {
		v :=  make(map[string]interface{})
		v["name"] = volume.Name
		switch {
			case volume.EmptyDir != nil:
				emptyDir := make(map[string]string)
				switch {
					case volume.EmptyDir.Medium == api.StorageMediumMemory:
						emptyDir["medium"] = "Memory"
					case volume.EmptyDir.Medium == api.StorageMediumDefault:
						emptyDir["medium"] = ""
				}
				v["emptyDir"] = emptyDir
			case volume.HostPath != nil:
				hostPath := make(map[string]string)
				hostPath["path"] = volume.HostPath.Path
				v["hostPath"] = hostPath
			case volume.AWSElasticBlockStore != nil:
				awsElasticBlockStore := make(map[string]interface{})
				awsElasticBlockStore["volumeID"]  = volume.AWSElasticBlockStore.VolumeID
				awsElasticBlockStore["fsType"]    = volume.AWSElasticBlockStore.FSType
				awsElasticBlockStore["partition"] = volume.AWSElasticBlockStore.Partition
				awsElasticBlockStore["readOnly"]  = volume.AWSElasticBlockStore.ReadOnly
				v["awsElasticBlockStore"] = awsElasticBlockStore
			default:
				panic("unknown volume type")
		}
		volumes = append(volumes, v)
	}
	d.Set("volume", volumes)

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

	var volumes []map[string]interface{}
	for _, volume := range pod.Spec.Volumes {
		v :=  make(map[string]interface{})
		v["name"] = volume.Name
		switch {
			case volume.EmptyDir != nil:
				emptyDir := make(map[string]string)
				switch {
					case volume.EmptyDir.Medium == api.StorageMediumMemory:
						emptyDir["medium"] = "Memory"
					case volume.EmptyDir.Medium == api.StorageMediumDefault:
						emptyDir["medium"] = ""
				}
				v["emptyDir"] = emptyDir
			case volume.HostPath != nil:
				hostPath := make(map[string]string)
				hostPath["path"] = volume.HostPath.Path
				v["hostPath"] = hostPath
			case volume.AWSElasticBlockStore != nil:
				awsElasticBlockStore := make(map[string]interface{})
				awsElasticBlockStore["volumeID"]  = volume.AWSElasticBlockStore.VolumeID
				awsElasticBlockStore["fsType"]    = volume.AWSElasticBlockStore.FSType
				awsElasticBlockStore["partition"] = volume.AWSElasticBlockStore.Partition
				awsElasticBlockStore["readOnly"]  = volume.AWSElasticBlockStore.ReadOnly
				v["awsElasticBlockStore"] = awsElasticBlockStore
			default:
				panic("unknown volume type")
		}
		volumes = append(volumes, v)
	}

	pod_map = make(map[string]interface{})
	pod_map["container"] = c_holder
	pod_map["volume"] = volumes
	pod_map["labels"] = pod.Labels
	pod_map["nodeName"] = pod.Spec.NodeName
	pod_map["namespace"] = pod.Namespace //currently not set because pods are inline to RC. handled on the RC side
	pod_map["name"] = pod.Name //this currently doesn't work properly because pods are still inline and don't define a name
	pod_map["terminationGracePeriodSeconds"] = strconv.FormatInt(*pod.Spec.TerminationGracePeriodSeconds,10)

	return pod_map, err
}
