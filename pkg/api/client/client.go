package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/metal-toolbox/component-inventory/pkg/api/constants"

	"github.com/metal-toolbox/alloy/types"
	rivets "github.com/metal-toolbox/rivets/types"
)

type ServerComponents map[string][]*rivets.Component

// Client can perform queries against the Component Inventory Service.
type Client interface {
	Version(context.Context) (string, error)
	GetServerComponents(context.Context, string, bool) (ServerComponents, error)
	UpdateInbandInventory(context.Context, string, *types.InventoryDevice) (string, error)
	UpdateOutOfbandInventory(context.Context, string, *types.InventoryDevice) (string, error)
}

type cisClient struct {
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
	client := cisClient{serverAddress: serverAddress}
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

func (c cisClient) GetServerComponents(ctx context.Context, serverID string, inband bool) (ServerComponents, error) {
	mode := constants.OutOfBandMode
	if inband {
		mode = constants.InBandMode
	}

	path := fmt.Sprintf("%v/%v?mode=%s", constants.ComponentsEndpoint, serverID, mode)
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

func (c cisClient) Version(ctx context.Context) (string, error) {
	resp, err := c.get(ctx, constants.VersionEndpoint)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

func (c cisClient) UpdateInbandInventory(ctx context.Context, serverID string, device *types.InventoryDevice) (string, error) {
	path := fmt.Sprintf("%v/%v?mode=inband", constants.InventoryEndpoint, serverID)
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

func (c cisClient) UpdateOutOfbandInventory(ctx context.Context, serverID string, device *types.InventoryDevice) (string, error) {
	path := fmt.Sprintf("%v/%v?mode=outofband", constants.InventoryEndpoint, serverID)
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
