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

// ContainerAppJobV0ToV1 migrates azurerm_container_app_job state from schema version 0 to 1.
// This follows the same pattern as ContainerAppV0ToV1 — see that type's documentation for
// details on why the migration exists and how the v0 schema overrides work.
type ContainerAppJobV0ToV1 struct{}

func (ContainerAppJobV0ToV1) Schema() map[string]*pluginsdk.Schema {
	schema := map[string]*pluginsdk.Schema{
		// --- shared properties (same order as container_app_v0_to_v1.go) ---

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

		"template": helpers.JobTemplateSchema(),

		"secret": helpers.SecretsSchema(),

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

		// --- container_app_job-specific properties ---

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

	// Override env from TypeSet back to TypeList for v0 schema
	if templateSchema, ok := schema["template"]; ok {
		if templateResource, ok := templateSchema.Elem.(*pluginsdk.Resource); ok {
			if containerSchema, ok := templateResource.Schema["container"]; ok {
				if containerResource, ok := containerSchema.Elem.(*pluginsdk.Resource); ok {
					if envSchema, ok := containerResource.Schema["env"]; ok {
						envSchema.Type = pluginsdk.TypeList
						envSchema.Set = nil
					}
				}
			}

			if initContainerSchema, ok := templateResource.Schema["init_container"]; ok {
				if initContainerResource, ok := initContainerSchema.Elem.(*pluginsdk.Resource); ok {
					if envSchema, ok := initContainerResource.Schema["env"]; ok {
						envSchema.Type = pluginsdk.TypeList
						envSchema.Set = nil
					}
				}
			}
		}
	}

	// Override secret set hash to nil for v0 schema
	if secretSchema, ok := schema["secret"]; ok {
		secretSchema.Set = nil
	}

	return schema
}

func (ContainerAppJobV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		return rawState, nil
	}
}
