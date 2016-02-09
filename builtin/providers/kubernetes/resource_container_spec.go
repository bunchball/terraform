package kubernetes

import (
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
	"strconv"
	"strings"
)

func resourceContainerSpec() map[string]*schema.Schema {
	s := make(map[string]*schema.Schema)

	s["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	}

	s["image"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}

	s["volumeMount"] = &schema.Schema{
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
				"readOnly": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
					Default:  "TCP",
					ForceNew: true,
				},
				"mountPath": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
			},
		},
	}

	s["env"] = &schema.Schema{
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
				"value": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "TCP",
					ForceNew: true,
				},
				//"valueFrom": &schema.Schema{ //this is complicated so will add it later
				//	Type:     schema.TypeString,
				//	Required: true,
				//	ForceNew: true,
				//},
			},
		},
	}

	s["port"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"protocol": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "TCP",
					ForceNew: true,
				},
				"containerPort": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"name": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
			},
		},
	}

	s["livenessProbe"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: resourceProbeSpec(),
		},
	}

	s["readinessProbe"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: resourceProbeSpec(),
		},
	}

	s["command"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem:     &schema.Schema {
			Type: schema.TypeString,
		},
	}

	s["args"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem:     &schema.Schema {
			Type: schema.TypeString,
		},
	}

	return s
}

func constructContainerSpec(c_tf_map map[string]interface{}) (c api.Container, err error) {
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

	env := c_tf_map["env"].([]interface{})
	for _, e_tf := range env {
		e_tf_map := e_tf.(map[string]interface{})

		var e api.EnvVar
		e.Name = e_tf_map["name"].(string)
		e.Value = e_tf_map["value"].(string)

		c.Env = append(c.Env, e)
	}

	vol := c_tf_map["volumeMount"].([]interface{})
	for _, v_tf := range vol {
		v_tf_map := v_tf.(map[string]interface{})

		var v api.VolumeMount
		v.Name = v_tf_map["name"].(string)
		v.ReadOnly = v_tf_map["readOnly"].(bool)
		v.MountPath = v_tf_map["mountPath"].(string)

		c.VolumeMounts = append(c.VolumeMounts, v)
	}

	if l_probe_i, ok := c_tf_map["livenessProbe"]; ok {
		if l_probe_list := l_probe_i.([]interface{}); len(l_probe_list) > 0 {
			l_probe_map := l_probe_list[0].(map[string]interface{})
			l_probe, e := constructProbeSpec(l_probe_map)
			err = e
			c.LivenessProbe = &l_probe
		}
	}

	if r_probe_i, ok := c_tf_map["readinessProbe"]; ok {
		if r_probe_list := r_probe_i.([]interface{}); len(r_probe_list) > 0 {
			r_probe_map := r_probe_list[0].(map[string]interface{})
			r_probe, e := constructProbeSpec(r_probe_map)
			err = e
			c.ReadinessProbe = &r_probe
		}
	}

	if command_list, ok := c_tf_map["command"]; ok {
		var c_list []string
		for _, c := range command_list.([]interface{}) {
			c_list = append(c_list, c.(string))
		}
		c.Command = c_list
	}

	if args_list, ok := c_tf_map["args"]; ok {
		var a_list []string
		for _, c := range args_list.([]interface{}) {
			a_list = append(a_list, c.(string))
		}
		c.Args = a_list
	}

	return c, err
}

func extractContainerSpec (c api.Container) (container map[string]interface{}, err error) {
	container = make(map[string]interface{})
	container["name"] = c.Name
	container["image"] = c.Image
	var portList []interface{}
	for _, p := range c.Ports {
		var portMap = make(map[string]interface{})
		portMap["name"] = p.Name
		portMap["containerPort"] = strconv.Itoa(p.ContainerPort)
		portMap["protocol"] = p.Protocol 
		portList = append(portList, portMap)
	}
	container["port"] = portList

	var envList []interface{}
	for _, e := range c.Env {
		var envMap = make(map[string]interface{})
		envMap["name"] = e.Name
		envMap["value"] = e.Value
		envList = append(envList, envMap)
	}
	container["env"] = envList

	var volList []interface{}
	for _, v := range c.VolumeMounts {
		var volMap = make(map[string]interface{})
		volMap["name"] = v.Name
		volMap["readOnly"] = v.ReadOnly
		volMap["mountPath"] = v.MountPath
		volList = append(volList, volMap)
	}
	container["volumeMount"] = volList

	if c.LivenessProbe != nil {
		l_probe, e := extractProbeSpec(c.LivenessProbe)
		err = e
		container["livenessProbe"] = l_probe
	}

	if c.ReadinessProbe != nil {
		r_probe, e := extractProbeSpec(c.ReadinessProbe)
		err = e
		container["readinessProbe"] = r_probe
	}

	if c.Command != nil {
		container["command"] = c.Command
	}

	if c.Args != nil {
		container["args"] = c.Args
	}

	return container, err
}
