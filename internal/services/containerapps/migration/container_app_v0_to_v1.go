// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package migration

import (
	"context"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/containerapps/helpers"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

// ContainerAppV0ToV1 migrates azurerm_container_app state from schema version 0 to 1.
//
// In v0, the env block was TypeList (index-based) and the secret set had no custom hash.
// In v1, env is TypeSet with containerEnvVarHash, and secret uses containerSecretHash.
//
// Schema() returns a static snapshot of the v0 schema — it reuses the current helper
// schemas (which are already v1/TypeSet) and overrides the env blocks back to TypeList
// and removes the secret set hash to match what v0 state looks like.
//
// UpgradeFunc() is a pass-through (returns rawState unchanged) because Terraform handles
// the TypeList-to-TypeSet and hash function changes automatically during state refresh.
// The migration exists solely to bump the schema version so Terraform knows a refresh
// is required.
type ContainerAppV0ToV1 struct{}

func (ContainerAppV0ToV1) Schema() map[string]*pluginsdk.Schema {
	schema := map[string]*pluginsdk.Schema{
		// --- shared properties (same order as container_app_job_v0_to_v1.go) ---

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

		"template": helpers.ContainerTemplateSchema(),

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

		// --- container_app-specific properties ---

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

	// Override env from TypeSet back to TypeList for v0 schema — in v0, env was
	// TypeList so the state has indexed keys (env.0.name, env.1.name, etc.) rather
	// than hash-based keys (env.12345.name). This must match the old state format.
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

	// Override secret set hash to nil for v0 schema — in v0, the secret set used the
	// SDK's default hash (all fields). Setting Set to nil restores that behavior.
	if secretSchema, ok := schema["secret"]; ok {
		secretSchema.Set = nil
	}

	return schema
}

func (ContainerAppV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		return rawState, nil
	}
}
