package kubernetes

import (
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api"
)

func resourceMeta() map[string]*schema.Schema {
       return map[string]*schema.Schema{
  	     	"name": &schema.Schema{
  	     		Type:     schema.TypeString,
  	     		Optional: true, //this really should be required, but the RC/PodTemplate crap is interfering. Will add the validation elsewhere?
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
       }
}

func constructMeta(d *schema.ResourceData) (meta api.ObjectMeta, err error) {
	
	return meta, err
}

func extractMeta(d *schema.ResourceData, meta *api.ObjectMeta) (err error) {

	return nil
}
