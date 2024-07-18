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

	tenantClient.BaseClient.Endpoint = "https://management.azure.com/subscriptions"
	//this endpoint is currently in public preview
	tenantClient.BaseClient.ApiVersion = msgraph.ManagementCIAMApiVersion

	return &Client{
		TenantClient: tenantClient,
	}
}
