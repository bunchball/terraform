package kubernetes

import (
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
	"strconv"
	"strings"
)

func resourceContainerSpec() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"image": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		"port": &schema.Schema{
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
		},
	}
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
	err = nil
	return c, err
}

func extractContainerSpec (v api.Container) (container map[string]interface{}, err error) {
	container = make(map[string]interface{})
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
	err = nil
	return container, err
}
