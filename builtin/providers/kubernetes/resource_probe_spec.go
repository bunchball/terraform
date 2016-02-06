package kubernetes

import (
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
	"strconv"
//	"reflect"
	"k8s.io/kubernetes/pkg/util"
	"strings"
	"regexp"
)

func resourceProbeSpec() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"initialDelaySeconds": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
		},
		"timeoutSeconds": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
		},
		"exec": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"command": &schema.Schema{
						Type:     schema.TypeList,
						Required: true,
						ForceNew: true,
						Elem:     &schema.Schema {
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"httpGet": &schema.Schema{
			Type:     schema.TypeMap,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"path": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
					"port": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
					"host": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
					"scheme": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
				},
			},
		},
	}
}

func constructProbeSpec(probe_map map[string]interface{}) (probe api.Probe, err error) {
	probe.InitialDelaySeconds = int64(probe_map["initialDelaySeconds"].(int))
	probe.TimeoutSeconds      = int64(probe_map["timeoutSeconds"].(int))
	
	var h api.Handler

	if exec_i, ok := probe_map["exec"]; ok {
		if len(exec_i.([]interface{})) > 0 {
			exec_map := exec_i.([]interface{})[0].(map[string]interface{})
			//exec_map := exec_i.(map[string]interface{})
			if exec_list, ok := exec_map["command"]; ok {
				var exec api.ExecAction

				var c_list []string
				for _, c := range exec_list.([]interface{}) {
					c_list = append(c_list, c.(string))
				}
				exec.Command = c_list
				h.Exec = &exec
			}
		}
	}

	if httpGet_i, ok := probe_map["httpGet"]; ok {
		httpGet_map := httpGet_i.(map[string]interface{})
		if len(httpGet_map) > 0 {
			var httpGet api.HTTPGetAction
			httpGet.Path = httpGet_map["path"].(string)


			port := httpGet_map["port"].(string)
			if number, _ := regexp.MatchString("[0-9]+", port); number {
				num_port, _ := strconv.Atoi(port)
				httpGet.Port = util.NewIntOrStringFromInt(num_port)
			} else if name, _  := regexp.MatchString("[a-z0-9]([a-z0-9-]*[a-z0-9])*", port); name {
				httpGet.Port = util.NewIntOrStringFromString(port)
			}

			if host, ok := httpGet_map["host"]; ok {
				httpGet.Host = host.(string)
			}

			if scheme_i, ok := httpGet_map["scheme"]; ok {
				switch scheme := strings.ToUpper(scheme_i.(string)); scheme {
					case "HTTP":
						httpGet.Scheme = api.URISchemeHTTP
					case "HTTPS":
						httpGet.Scheme = api.URISchemeHTTPS
					default:
						panic("not a valid scheme")
				}
			}

			h.HTTPGet = &httpGet
		}
	}
	probe.Handler = h

	return probe, err
}

func extractProbeSpec (probe *api.Probe) (probe_map map[string]interface{}, err error) {
	probe_map = make(map[string]interface{})
	probe_map["initialDelaySeconds"] = probe.InitialDelaySeconds
	probe_map["timeoutSeconds"] = probe.TimeoutSeconds

	h := probe.Handler
	switch {
		case h.Exec != nil:
			exec_list := h.Exec.Command
			probe_map["exec"] = exec_list
		case h.HTTPGet != nil:
			httpGet_map := make(map[string]interface{})
			httpGet_map["path"] = h.HTTPGet.Path
			httpGet_map["port"] = h.HTTPGet.Port
			httpGet_map["host"] = h.HTTPGet.Host
			switch h.HTTPGet.Scheme {
				case api.URISchemeHTTP:
					httpGet_map["scheme"] = "HTTP"
				case api.URISchemeHTTPS:
					httpGet_map["scheme"] = "HTTPS"
			}
	}

	return probe_map, err
}
