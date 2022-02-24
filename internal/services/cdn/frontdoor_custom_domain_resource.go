package cdn

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/cdn/mgmt/2020-09-01/cdn"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cdn/parse"
	track1 "github.com/hashicorp/terraform-provider-azurerm/internal/services/cdn/sdk/2021-06-01"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cdn/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceFrontdoorCustomDomain() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceFrontdoorCustomDomainCreate,
		Read:   resourceFrontdoorCustomDomainRead,
		Update: resourceFrontdoorCustomDomainUpdate,
		Delete: resourceFrontdoorCustomDomainDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.FrontdoorCustomDomainID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"frontdoor_profile_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ProfileID,
			},

			"azure_dns_zone_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"domain_validation_state": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"host_name": {
				Type:     pluginsdk.TypeString,
				ForceNew: true,
				Required: true,
			},

			"pre_validated_custom_domain_resource_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"profile_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tls_settings": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,

				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{

						"certificate_type": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							Default:  string(cdn.AfdCertificateTypeCustomerCertificate),
							ValidateFunc: validation.StringInSlice([]string{
								string(cdn.AfdCertificateTypeCustomerCertificate),
								string(cdn.AfdCertificateTypeManagedCertificate),
							}, false),
						},

						"minimum_tls_version": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							Default:  string(cdn.AfdMinimumTLSVersionTLS12),
							ValidateFunc: validation.StringInSlice([]string{
								string(cdn.AfdMinimumTLSVersionTLS10),
								string(cdn.AfdMinimumTLSVersionTLS12),
							}, false),
						},

						"secret_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
					},
				},
			},

			"validation_properties": {
				Type:     pluginsdk.TypeList,
				Computed: true,

				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{

						"expiration_date": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"validation_token": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceFrontdoorCustomDomainCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cdn.FrontDoorCustomDomainsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	profileId, err := parse.FrontdoorProfileID(d.Get("frontdoor_profile_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewFrontdoorCustomDomainID(profileId.SubscriptionId, profileId.ResourceGroup, profileId.ProfileName, d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.ProfileName, id.CustomDomainName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_frontdoor_custom_domain", id.ID())
		}
	}

	props := track1.AFDDomain{
		AFDDomainProperties: &track1.AFDDomainProperties{
			AzureDNSZone:                       expandCustomDomainResourceReference(d.Get("azure_dns_zone").(string)),
			PreValidatedCustomDomainResourceID: expandCustomDomainResourceReference(d.Get("pre_validated_custom_domain_resource_id").(string)),
			TLSSettings:                        expandCustomDomainAFDDomainHttpsParameters(d.Get("tls_settings").([]interface{})),
		},
	}

	future, err := client.Create(ctx, id.ResourceGroup, id.ProfileName, id.CustomDomainName, props)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the creation of %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceFrontdoorCustomDomainRead(d, meta)
}

func resourceFrontdoorCustomDomainRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cdn.FrontDoorCustomDomainsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FrontdoorCustomDomainID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ProfileName, id.CustomDomainName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.CustomDomainName)

	d.Set("frontdoor_profile_id", parse.NewFrontdoorProfileID(id.SubscriptionId, id.ResourceGroup, id.ProfileName).ID())

	resp, err = client.Get(ctx, id.ResourceGroup, id.ProfileName, id.CustomDomainName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("making Read request on Azure CDN Profile %q (Resource Group %q): %+v", id.ProfileName, id.ResourceGroup, err)
	}

	if props := resp.AFDDomainProperties; props != nil {
		d.Set("domain_validation_state", props.DomainValidationState)
		d.Set("host_name", props.HostName)
		d.Set("frontdoor_profile_name", props.ProfileName)

		if err := d.Set("azure_dns_zone", flattenCustomDomainResourceReference(props.AzureDNSZone)); err != nil {
			return fmt.Errorf("setting `azure_dns_zone`: %+v", err)
		}

		if err := d.Set("pre_validated_custom_domain_resource_id", flattenCustomDomainResourceReference(props.PreValidatedCustomDomainResourceID)); err != nil {
			return fmt.Errorf("setting `pre_validated_custom_domain_resource_id`: %+v", err)
		}

		if err := d.Set("tls_settings", flattenCustomDomainAFDDomainHttpsParameters(props.TLSSettings)); err != nil {
			return fmt.Errorf("setting `tls_settings`: %+v", err)
		}

		if err := d.Set("validation_properties", flattenCustomDomainDomainValidationProperties(props.ValidationProperties)); err != nil {
			return fmt.Errorf("setting `validation_properties`: %+v", err)
		}
	}

	return nil
}

func resourceFrontdoorCustomDomainUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cdn.FrontDoorCustomDomainsClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FrontdoorCustomDomainID(d.Id())
	if err != nil {
		return err
	}

	props := track1.AFDDomainUpdateParameters{
		AFDDomainUpdatePropertiesParameters: &track1.AFDDomainUpdatePropertiesParameters{
			AzureDNSZone:                       expandCustomDomainResourceReference(d.Get("azure_dns_zone").([]interface{})),
			PreValidatedCustomDomainResourceID: expandCustomDomainResourceReference(d.Get("pre_validated_custom_domain_resource_id").([]interface{})),
			TLSSettings:                        expandCustomDomainAFDDomainHttpsParameters(d.Get("tls_settings").([]interface{})),
		},
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.ProfileName, id.CustomDomainName, props)
	if err != nil {
		return fmt.Errorf("updating %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the update of %s: %+v", *id, err)
	}

	return resourceFrontdoorCustomDomainRead(d, meta)
}

func resourceFrontdoorCustomDomainDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cdn.FrontDoorCustomDomainsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FrontdoorCustomDomainID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.ProfileName, id.CustomDomainName)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of %s: %+v", *id, err)
	}

	return err
}

func expandCustomDomainResourceReference(input interface{}) *track1.ResourceReference {
	if input == nil {
		return nil
	}

	return &track1.ResourceReference{
		ID: utils.String(input.(string)),
	}
}

func expandCustomDomainAFDDomainHttpsParameters(input []interface{}) *track1.AFDDomainHTTPSParameters {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	v := input[0].(map[string]interface{})

	certificateTypeValue := track1.AfdCertificateType(v["certificate_type"].(string))
	minimumTlsVersionValue := track1.AfdMinimumTLSVersion(v["minimum_tls_version"].(string))
	return &track1.AFDDomainHTTPSParameters{
		CertificateType:   certificateTypeValue,
		MinimumTLSVersion: minimumTlsVersionValue,
		Secret:            expandCustomDomainResourceReference(v["secret_id"].(string)),
	}
}

func flattenCustomDomainAFDDomainHttpsParameters(input *track1.AFDDomainHTTPSParameters) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	result := make(map[string]interface{})
	result["certificate_type"] = input.CertificateType
	result["minimum_tls_version"] = input.MinimumTLSVersion

	result["secret_id"] = *flattenCustomDomainResourceReference(input.Secret)
	return append(results, result)
}

func flattenCustomDomainDomainValidationProperties(input *track1.DomainValidationProperties) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	result := make(map[string]interface{})

	if input.ExpirationDate != nil {
		result["expiration_date"] = *input.ExpirationDate
	}

	if input.ValidationToken != nil {
		result["validation_token"] = *input.ValidationToken
	}
	return append(results, result)
}

func flattenCustomDomainResourceReference(input *track1.ResourceReference) *string {
	if input == nil {
		return nil
	}

	return input.ID
}