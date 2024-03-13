package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/component-inventory/pkg/api/routes"
	rivets "github.com/metal-toolbox/rivets/types"
)

type ServerComponents map[string][]*rivets.Component

// Client can perform queries against the Component Inventory Service.
type Client interface {
	Version(context.Context) (string, error)
	GetServerComponents(context.Context, string) (ServerComponents, error)
	UpdateInbandInventory(context.Context, string, *common.Device) (string, error)
	UpdateOutOfbandInventory(context.Context, string, *common.Device) (string, error)
}

type componentInventoryClient struct {
	// The server address with the schema
	serverAddress string
	// Authentication token
	authToken string
	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	client httpRequestDoer
}

// Creates a new Client, with reasonable defaults
func NewClient(serverAddress string, opts ...Option) (Client, error) {
	// create a client with sane default values
	client := componentInventoryClient{serverAddress: serverAddress}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}

	// create httpClient, if not already present
	if client.client == nil {
		client.client = &http.Client{}
	}

	return client, nil
}

func (c componentInventoryClient) GetServerComponents(ctx context.Context, serverID string) (ServerComponents, error) {
	path := fmt.Sprintf("%v/%v", routes.ComponentsEndpoint, serverID)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	sc := make(map[string][]*rivets.Component)
	err = json.Unmarshal(resp, &sc)
	if err != nil {
		return nil, err
	}
	return sc, err
}

func (c componentInventoryClient) Version(ctx context.Context) (string, error) {
	resp, err := c.get(ctx, routes.VersionEndpoint)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

func (c componentInventoryClient) UpdateInbandInventory(ctx context.Context, serverID string, device *common.Device) (string, error) {
	path := fmt.Sprintf("%v/%v", routes.InbandInventoryEndpoint, serverID)
	body, err := json.Marshal(device)
	if err != nil {
		return "", fmt.Errorf("failed to parse device: %v", err)
	}

	resp, err := c.post(ctx, path, body)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

func (c componentInventoryClient) UpdateOutOfbandInventory(ctx context.Context, serverID string, device *common.Device) (string, error) {
	path := fmt.Sprintf("%v/%v", routes.OutofbandInventoryEndpoint, serverID)
	body, err := json.Marshal(device)
	if err != nil {
		return "", fmt.Errorf("failed to parse device: %v", err)
	}

	resp, err := c.post(ctx, path, body)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}
