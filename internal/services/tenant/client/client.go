package tenant

import (
	"github.com/hashicorp/terraform-provider-azuread/internal/common"
	"github.com/manicminer/hamilton/msgraph"
)

type Client struct {
	TenantClient *msgraph.TenantClient
}

func NewClient(o *common.ClientOptions) *Client {

	tenantClient := msgraph.NewTenantClient()
	o.ConfigureClient(&tenantClient.BaseClient)

	tenantClient.BaseClient.Endpoint = "https://management.azure.com"
	/*
		ugly workaround but since we are not hitting the standard graph endpoint
		we need to set the api version here, which in turn allows it to use the correct base url.
		The Base Client might need some overrides if this is not a preferable solution
	*/
	tenantClient.BaseClient.ApiVersion = "subscriptions"

	return &Client{
		TenantClient: tenantClient,
	}
}
