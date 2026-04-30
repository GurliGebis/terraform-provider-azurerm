// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package migration

import (
	"context"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/containerapps/helpers"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type ContainerAppV0ToV1 struct{}

func envSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		MinItems: 1,
		Optional: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"name": {
					Type:     pluginsdk.TypeString,
					Required: true,
				},
				"value": {
					Type:     pluginsdk.TypeString,
					Optional: true,
				},
				"secret_name": {
					Type:     pluginsdk.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func secretSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:      pluginsdk.TypeSet,
		Optional:  true,
		Sensitive: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"identity": {
					Type:     pluginsdk.TypeString,
					Optional: true,
				},
				"key_vault_secret_id": {
					Type:     pluginsdk.TypeString,
					Optional: true,
				},
				"name": {
					Type:     pluginsdk.TypeString,
					Required: true,
				},
				"value": {
					Type:      pluginsdk.TypeString,
					Optional:  true,
					Sensitive: true,
				},
			},
		},
	}
}

func (ContainerAppV0ToV1) Schema() map[string]*pluginsdk.Schema {
	templateSchema := helpers.ContainerTemplateSchema()

	if templateResource, ok := templateSchema.Elem.(*pluginsdk.Resource); ok {
		if containerSchema, ok := templateResource.Schema["container"]; ok {
			if containerResource, ok := containerSchema.Elem.(*pluginsdk.Resource); ok {
				containerResource.Schema["env"] = envSchemaV0()
			}
		}
		if initContainerSchema, ok := templateResource.Schema["init_container"]; ok {
			if initContainerResource, ok := initContainerSchema.Elem.(*pluginsdk.Resource); ok {
				initContainerResource.Schema["env"] = envSchemaV0()
			}
		}
	}

	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"resource_group_name": commonschema.ResourceGroupName(),

		"location": commonschema.LocationComputed(),

		"container_app_environment_id": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"workload_profile_name": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"template": templateSchema,

		"secret": secretSchemaV0(),

		"registry": helpers.ContainerAppRegistrySchema(),

		"identity": commonschema.SystemAssignedUserAssignedIdentityOptional(),

		"tags": commonschema.Tags(),

		"outbound_ip_addresses": {
			Type:     pluginsdk.TypeList,
			Computed: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"id": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Computed: true,
		},

		"revision_mode": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"ingress": helpers.ContainerAppIngressSchema(),

		"dapr": helpers.ContainerDaprSchema(),

		"max_inactive_revisions": {
			Type:     pluginsdk.TypeInt,
			Optional: true,
		},

		"latest_revision_name": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"latest_revision_fqdn": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"custom_domain_verification_id": {
			Type:      pluginsdk.TypeString,
			Computed:  true,
			Sensitive: true,
		},
	}
}

func (ContainerAppV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		normalizeEnvVars(rawState)
		normalizeSecrets(rawState)
		return rawState, nil
	}
}

func normalizeEnvVars(rawState map[string]interface{}) {
	templates, _ := rawState["template"].([]interface{})
	for _, tmpl := range templates {
		tmplMap, ok := tmpl.(map[string]interface{})
		if !ok {
			continue
		}
		for _, key := range []string{"container", "init_container"} {
			containers, _ := tmplMap[key].([]interface{})
			for _, c := range containers {
				cMap, ok := c.(map[string]interface{})
				if !ok {
					continue
				}
				envs, _ := cMap["env"].([]interface{})
				for _, e := range envs {
					eMap, ok := e.(map[string]interface{})
					if !ok {
						continue
					}
					if eMap["value"] == nil {
						eMap["value"] = ""
					}
					if eMap["secret_name"] == nil {
						eMap["secret_name"] = ""
					}
				}
			}
		}
	}
}

func normalizeSecrets(rawState map[string]interface{}) {
	secrets, _ := rawState["secret"].([]interface{})
	for _, s := range secrets {
		sMap, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		if sMap["identity"] == nil {
			sMap["identity"] = ""
		}
		if sMap["key_vault_secret_id"] == nil {
			sMap["key_vault_secret_id"] = ""
		}
		if sMap["value"] == nil {
			sMap["value"] = ""
		}
	}
}
