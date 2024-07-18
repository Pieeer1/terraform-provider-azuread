package tenant

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azuread/internal/clients"
	"github.com/hashicorp/terraform-provider-azuread/internal/tf"
	"github.com/hashicorp/terraform-provider-azuread/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azuread/internal/tf/validation"
	"github.com/manicminer/hamilton/msgraph"
)

func tenantResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: tenantResourceCreate,
		ReadContext:   tenantResourceRead,
		UpdateContext: tenantResourceUpdate,
		DeleteContext: tenantResourceDelete,
		Schema: map[string]*pluginsdk.Schema{
			"resource_group_name": {
				Description:      "The name of the resource group in which the child tenant should be created",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ValidateDiag(validation.StringIsNotWhiteSpace),
			},
			"domain_name": {
				Description: "The unique alpha-numeric domain name of the child tenant",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"subscription_id": {
				Description:      "The subscription ID of the resource group tenant",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ValidateDiag(validation.IsUUID),
			},
			"location": {
				Description:      "The country of the child tenant. Possible Values: [United States, Eurpose, Asia Pacific, Australia]",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ValidateDiag(validation.StringIsNotWhiteSpace),
			},
			"sku_name": {
				Description:      "The SKU name of the child tenant. Possible Values: [Base]",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ValidateDiag(validation.StringIsNotWhiteSpace),
			},
			"display_name": {
				Description:      "The display name of the child tenant",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ValidateDiag(validation.StringIsNotWhiteSpace),
			},
			"api_version": {
				Description: "The API version of the Azure Resource Manager",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "2023-05-17-preview",
			},
			"tenant_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The tenant ID of the ExternalEntra Tenant",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The tags of the ExternalEntra Tenant",
			},
		},
	}
}

func tenantResourceCreate(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) pluginsdk.Diagnostics {
	client := meta.(*clients.Client).Tenant.TenantClient

	tenant, err := createTenant(d)
	if err != nil {
		return tf.ErrorDiagPathF(err, "location", "Invalid country code")
	}

	respTenant, _, err := client.Create(ctx, *tenant, d.Get("domain_name").(string), d.Get("subscription_id").(string), d.Get("resource_group_name").(string))

	if err != nil {
		return tf.ErrorDiagPathF(err, "domain_name", "Error creating tenant: %v", err)
	}

	setTenantValues(d, respTenant)

	return nil
}

func tenantResourceRead(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) pluginsdk.Diagnostics {
	client := meta.(*clients.Client).Tenant.TenantClient

	tenant, _, err := client.Get(ctx, d.Get("domain_name").(string), d.Get("subscription_id").(string), d.Get("resource_group_name").(string))
	if err != nil {
		return tf.ErrorDiagPathF(err, "domain_name", "Error reading tenant: %v", err)
	}

	setTenantValues(d, tenant)

	return nil
}

func tenantResourceUpdate(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) pluginsdk.Diagnostics {
	client := meta.(*clients.Client).Tenant.TenantClient

	tenant, err := createTenant(d)
	if err != nil {
		return tf.ErrorDiagPathF(err, "location", "Invalid country code")
	}

	respTenant, _, err := client.Update(ctx, *tenant, d.Get("domain_name").(string), d.Get("subscription_id").(string), d.Get("resource_group_name").(string))

	if err != nil {
		return tf.ErrorDiagPathF(err, "domain_name", "Error updating tenant: %v", err)
	}

	setTenantValues(d, respTenant)

	return nil
}

func tenantResourceDelete(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) pluginsdk.Diagnostics {
	client := meta.(*clients.Client).Tenant.TenantClient

	_, err := client.Delete(ctx, d.Get("domain_name").(string), d.Get("subscription_id").(string), d.Get("resource_group_name").(string))

	if err != nil {
		return tf.ErrorDiagPathF(err, "domain_name", "Error deleting tenant: %v", err)
	}

	d.SetId("")
	return nil
}

func setTenantValues(d *pluginsdk.ResourceData, tenant *msgraph.Tenant) {
	d.SetId(*tenant.Id)
	d.Set("tenant_id", *tenant.Properties.TenantId)
	d.Set("domain_name", *tenant.Properties.DomainName)
	d.Set("sku_name", *tenant.Sku.Name)
	d.Set("display_name", *tenant.Properties.CreateTenantProperties.DisplayName)
	d.Set("location", *tenant.Location)
	d.Set("tags", *tenant.Tags)
}

func determineCountryCode(country string) (string, error) {
	switch country {
	case "United States":
		return "US", nil
	case "Europe":
		return "EU", nil
	case "Asia Pacific":
		return "AP", nil
	case "Australia":
		return "AU", nil
	default:
		return "", fmt.Errorf("Invalid country code: %s", country)
	}

}

func createTenant(d *pluginsdk.ResourceData) (*msgraph.Tenant, error) {
	baseTier := "A0"
	country := d.Get("location").(*string)
	countryCode, err := determineCountryCode(*country)
	if err != nil {
		return nil, fmt.Errorf("invalid country code: %s", *country)
	}

	return &msgraph.Tenant{
		Location: country,
		Sku: &msgraph.TenantSku{
			Name: d.Get("sku_name").(*string),
			Tier: &baseTier,
		},
		Properties: &msgraph.TenantProperties{
			CreateTenantProperties: &msgraph.CreateTenantProperties{
				DisplayName: d.Get("display_name").(*string),
				CountryCode: &countryCode,
			},
		},
		Tags: d.Get("tags").(*map[string]string),
	}, nil

}
