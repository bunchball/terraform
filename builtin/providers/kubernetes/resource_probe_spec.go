package kubernetes

import (
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
//	"strconv"
	"reflect"
	"k8s.io/kubernetes/pkg/util"
	"strings"
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
			Type:     schema.TypeMap,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"command": &schema.Schema{
						Type:     schema.TypeList,
						Required: true,
						ForceNew: true,
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
	probe.InitialDelaySeconds = probe_map["initialDelaySeconds"].(int64)
	probe.TimeoutSeconds      = probe_map["timeoutSeconds"].(int64)
	
	var h api.Handler

	if exec_i, ok := probe_map["exec"]; ok {
		exec_list := exec_i.([]string)
		var exec api.ExecAction
		exec.Command = exec_list
		h.Exec = &exec
	}

	if httpGet_i, ok := probe_map["httpGet"]; ok {
		httpGet_map := httpGet_i.(map[string]interface{})
		var httpGet api.HTTPGetAction
		httpGet.Path = httpGet_map["path"].(string)

		switch typ := reflect.TypeOf(httpGet_map["port"]).Kind(); typ {
			case reflect.String:
				httpGet.Port = util.NewIntOrStringFromString(httpGet_map["port"].(string))
			case reflect.Int:
				httpGet.Port = util.NewIntOrStringFromInt(httpGet_map["port"].(int))
			default:
				panic("Not a string or int")	
		}

		httpGet.Host = httpGet_map["host"].(string)

		switch scheme := strings.ToUpper(httpGet_map["scheme"].(string)); scheme {
			case "HTTP":
				httpGet.Scheme = api.URISchemeHTTP
			case "HTTPS":
				httpGet.Scheme = api.URISchemeHTTPS
			default:
				panic("not a valid scheme")
		}


		h.HTTPGet = &httpGet
	}
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
