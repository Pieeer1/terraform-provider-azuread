package msgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TenantClient struct {
	BaseClient Client
}

const (
	// most recent version to date. This is the default version for the client.
	ManagementCIAMApiVersion = "2023-05-17-preview"
)

func NewTenantClient() *TenantClient {

	baseClient := NewClient(ManagementCIAMApiVersion)

	baseClient.Endpoint = "https://management.azure.com/subscriptions"

	return &TenantClient{
		baseClient,
	}
}

func (c *TenantClient) Get(ctx context.Context, domain string, subscriptionId string, resourceGroupName string) (*Tenant, int, error) {
	resp, status, _, err := c.BaseClient.Get(ctx, GetHttpRequestInput{
		ValidStatusCodes: []int{http.StatusOK},
		Uri: Uri{
			Entity: fmt.Sprintf("/%s/resourceGroups/%s/providers/Microsoft.AzureActiveDirectory/ciamDirectories/%s?api-version=%s", subscriptionId, resourceGroupName, domain, c.BaseClient.ApiVersion),
		},
	})
	if err != nil {
		return nil, status, fmt.Errorf("TenantClient.BaseClient.Get(): %v", err)
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status, fmt.Errorf("io.ReadAll(): %v", err)
	}

	var tenant Tenant
	err = json.Unmarshal(respBody, &tenant)

	if err != nil {
		return nil, status, fmt.Errorf("json.Unmarshal(): %v", err)
	}

	return &tenant, status, nil
}

func (c *TenantClient) Create(ctx context.Context, tenant Tenant, domain string, subscriptionId string, resourceGroupName string) (*Tenant, int, error) {
	var status int

	body, err := json.Marshal(tenant)
	if err != nil {
		return nil, status, fmt.Errorf("json.Marshal(): %v", err)
	}

	resp, status, _, err := c.BaseClient.Put(ctx, PutHttpRequestInput{
		Body:             body,
		ValidStatusCodes: []int{http.StatusCreated},
		Uri: Uri{
			Entity: fmt.Sprintf("/%s/resourceGroups/%s/providers/Microsoft.AzureActiveDirectory/ciamDirectories/%s?api-version=%s", subscriptionId, resourceGroupName, domain, c.BaseClient.ApiVersion),
		},
	})
	if err != nil {
		return nil, status, fmt.Errorf("TenantClient.BaseClient.Put(): %v", err)
	}

	defer resp.Body.Close()

	resultEndpoint := resp.Header.Get("Azure-AsyncOperation")
	resultWait, err := strconv.ParseFloat(resp.Header.Get("Retry-After"), 64)

	if err != nil {
		return nil, status, fmt.Errorf("strconv.ParseFloat(): %v", err)
	}

	status, err = c.poll(ctx, resultEndpoint, resultWait)

	if err != nil {
		return nil, status, fmt.Errorf("poll(): %v", err)
	}

	return c.Get(ctx, domain, subscriptionId, resourceGroupName)
}

func (c *TenantClient) poll(ctx context.Context, endpoint string, nextRequestIn float64) (int, error) {
	time.Sleep(time.Duration(nextRequestIn) * time.Second)

	resp, status, _, err := c.BaseClient.Get(ctx, GetHttpRequestInput{
		ValidStatusCodes: []int{http.StatusOK},
		Uri: Uri{
			Entity: strings.Split(endpoint, c.BaseClient.Endpoint)[1],
		},
	})
	if err != nil {
		return status, fmt.Errorf("TenantClient.BaseClient.Get(): %v", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return status, fmt.Errorf("io.ReadAll(): %v", err)
	}

	var responseBody map[string]interface{}

	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return status, fmt.Errorf("json.Unmarshal(): %v", err)
	}

	if responseBody["status"] == "Pending" {
		return c.poll(ctx, endpoint, nextRequestIn)
	} else if responseBody["status"] == "Succeeded" {
		return status, nil
	} else {
		return status, fmt.Errorf("poll(): %v", responseBody)
	}

}
