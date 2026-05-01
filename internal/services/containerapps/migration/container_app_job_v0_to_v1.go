// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package migration

import (
	"context"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/containerapps/helpers"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

type ContainerAppJobV0ToV1 struct{}

// jobEnvSchemaV0 is a point-in-time snapshot of the job env schema before v1.
// Must remain static: TypeList, MinItems:1, no Default.
func jobEnvSchemaV0() *pluginsdk.Schema {
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

// jobSecretSchemaV0 is a point-in-time snapshot of the job secret schema before v1.
// Key differences from v1: set-level Sensitive:true, no Default:"" on optional fields.
func jobSecretSchemaV0() *pluginsdk.Schema {
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

func (ContainerAppJobV0ToV1) Schema() map[string]*pluginsdk.Schema {
	templateSchema := helpers.JobTemplateSchema()

	if templateResource, ok := templateSchema.Elem.(*pluginsdk.Resource); ok {
		if containerSchema, ok := templateResource.Schema["container"]; ok {
			if containerResource, ok := containerSchema.Elem.(*pluginsdk.Resource); ok {
				containerResource.Schema["env"] = jobEnvSchemaV0()
			}
		}
		if initContainerSchema, ok := templateResource.Schema["init_container"]; ok {
			if initContainerResource, ok := initContainerSchema.Elem.(*pluginsdk.Resource); ok {
				initContainerResource.Schema["env"] = jobEnvSchemaV0()
			}
		}
	}

	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"resource_group_name": commonschema.ResourceGroupName(),

		"location": commonschema.Location(),

		"container_app_environment_id": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"workload_profile_name": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"template": templateSchema,

		"secret": jobSecretSchemaV0(),

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

		"replica_timeout_in_seconds": {
			Type:     pluginsdk.TypeInt,
			Required: true,
		},

		"replica_retry_limit": {
			Type:     pluginsdk.TypeInt,
			Optional: true,
		},

		"event_trigger_config": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"parallelism": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
					"replica_completion_count": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
					"scale": helpers.ContainerAppsJobsScaleSchema(),
				},
			},
		},

		"schedule_trigger_config": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"cron_expression": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},
					"parallelism": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
					"replica_completion_count": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
				},
			},
		},

		"manual_trigger_config": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"parallelism": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
					"replica_completion_count": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
				},
			},
		},

		"event_stream_endpoint": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}

func (ContainerAppJobV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		normalizeJobEnvVars(rawState)
		normalizeJobSecrets(rawState)
		return rawState, nil
	}
}

// normalizeJobEnvVars converts nil values to "" for env var fields that gained Default:"" in v1.
func normalizeJobEnvVars(rawState map[string]interface{}) {
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

// normalizeJobSecrets converts nil values to "" for secret fields that gained Default:"" in v1.
func normalizeJobSecrets(rawState map[string]interface{}) {
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
