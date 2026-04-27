// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-sdk/resource-manager/containerapps/2025-07-01/containerapps"
)

func TestValidateContainerAppRegistry(t *testing.T) {
	cases := []struct {
		Input Registry
		Valid bool
	}{
		{
			Input: Registry{
				Server:            "registry.example.com",
				UserName:          "user",
				PasswordSecretRef: "secretref",
			},
			Valid: true,
		},
		{
			Input: Registry{
				Server:   "registry.example.com",
				Identity: "identity",
			},
			Valid: true,
		},
		{
			Input: Registry{
				Server: "registry.example.com",
			},
			Valid: false,
		},
		{
			Input: Registry{
				Server:            "registry.example.com",
				UserName:          "user",
				PasswordSecretRef: "secretref",
				Identity:          "identity",
			},
			Valid: false,
		},
		{
			Input: Registry{
				Server:            "registry.example.com",
				PasswordSecretRef: "secretref",
			},
			Valid: false,
		},
		{
			Input: Registry{
				Server:   "registry.example.com",
				UserName: "user",
			},
			Valid: false,
		},
	}

	for _, tc := range cases {
		t.Logf("[DEBUG] Testing Value %s", tc.Input)
		err := ValidateContainerAppRegistry(tc.Input)
		valid := err == nil
		if tc.Valid != valid {
			t.Fatalf("Expected %t but got %t for %s", tc.Valid, valid, tc.Input)
		}
	}
}

func TestContainerEnvVarHashIdenticalEntries(t *testing.T) {
	a := map[string]interface{}{"name": "APP_MODE", "value": "blue"}
	b := map[string]interface{}{"name": "APP_MODE", "value": "blue"}
	if containerEnvVarHash(a) != containerEnvVarHash(b) {
		t.Fatal("expected equal hashes for identical entries")
	}
}

func TestContainerEnvVarHashDifferentValues(t *testing.T) {
	a := map[string]interface{}{"name": "APP_MODE", "value": "blue"}
	b := map[string]interface{}{"name": "APP_MODE", "value": "green"}
	if containerEnvVarHash(a) == containerEnvVarHash(b) {
		t.Fatal("expected different hashes for different values")
	}
}

func TestContainerEnvVarHashValueVsSecret(t *testing.T) {
	a := map[string]interface{}{"name": "APP_MODE", "value": "blue"}
	b := map[string]interface{}{"name": "APP_MODE", "secret_name": "s"}
	if containerEnvVarHash(a) == containerEnvVarHash(b) {
		t.Fatal("expected different hashes for value-backed vs secret-backed")
	}
}

func TestContainerEnvVarHashDifferentSecretNames(t *testing.T) {
	a := map[string]interface{}{"name": "APP_MODE", "secret_name": "a"}
	b := map[string]interface{}{"name": "APP_MODE", "secret_name": "b"}
	if containerEnvVarHash(a) == containerEnvVarHash(b) {
		t.Fatal("expected different hashes for different secret_name values")
	}
}

func TestContainerEnvVarHashDifferentNames(t *testing.T) {
	a := map[string]interface{}{"name": "APP_MODE", "value": "x"}
	b := map[string]interface{}{"name": "API_URL", "value": "x"}
	if containerEnvVarHash(a) == containerEnvVarHash(b) {
		t.Fatal("expected different hashes for different names")
	}
}

func TestContainerEnvVarHashCaseSensitive(t *testing.T) {
	a := map[string]interface{}{"name": "APP_MODE", "value": "x"}
	b := map[string]interface{}{"name": "app_mode", "value": "x"}
	if containerEnvVarHash(a) == containerEnvVarHash(b) {
		t.Fatal("expected different hashes for different name casing")
	}
}

func TestContainerSecretHashSameNameDifferentValue(t *testing.T) {
	a := map[string]interface{}{"name": "db-password", "value": "v1"}
	b := map[string]interface{}{"name": "db-password", "value": "v2"}
	if containerSecretHash(a) != containerSecretHash(b) {
		t.Fatal("expected equal hashes — secret identity is name-only")
	}
}

func TestContainerSecretHashDifferentNames(t *testing.T) {
	a := map[string]interface{}{"name": "db-password"}
	b := map[string]interface{}{"name": "api-key"}
	if containerSecretHash(a) == containerSecretHash(b) {
		t.Fatal("expected different hashes for different secret names")
	}
}

func TestFlattenContainerEnvVarFiltersEmptyNames(t *testing.T) {
	result := flattenContainerEnvVar(&[]containerapps.EnvironmentVar{
		{Name: pointer.To(""), SecretRef: pointer.To("s")},
		{Name: pointer.To("VALID"), SecretRef: pointer.To("s")},
	})
	if len(result) != 1 || result[0].Name != "VALID" {
		t.Fatalf("expected 1 env var named VALID, got %+v", result)
	}
}

func TestFlattenContainerAppSecretsNormalizesBlankKeyVaultURL(t *testing.T) {
	result := FlattenContainerAppSecrets(&containerapps.SecretsCollection{
		Value: []containerapps.ContainerAppSecret{
			{Name: pointer.To("s"), KeyVaultURL: pointer.To(""), Value: pointer.To("val")},
		},
	})
	if len(result) != 1 || result[0].KeyVaultSecretId != "" || result[0].Value != "val" {
		t.Fatalf("expected value-backed secret with blank key_vault_secret_id, got %+v", result)
	}
}

func TestPreserveContainerAppSecretValues(t *testing.T) {
	current := []Secret{
		{Name: "val-secret", Value: ""},
		{Name: "kv-secret", KeyVaultSecretId: "https://vault/secrets/kv", Value: ""},
	}
	prior := []Secret{
		{Name: "val-secret", Value: "preserved"},
		{Name: "kv-secret", KeyVaultSecretId: "https://vault/secrets/kv", Value: "ignored"},
	}

	result := PreserveContainerAppSecretValues(current, prior)

	if result[0].Value != "preserved" {
		t.Fatalf("expected value-backed secret to be preserved, got %q", result[0].Value)
	}
	if result[1].Value != "" {
		t.Fatalf("expected key-vault-backed secret value to stay empty, got %q", result[1].Value)
	}
}
