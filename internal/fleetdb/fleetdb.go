package internalfleetdb

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/metal-toolbox/component-inventory/internal/app"
	"github.com/metal-toolbox/component-inventory/internal/inventoryconverter"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	rivets "github.com/metal-toolbox/rivets/types"
	"go.uber.org/zap"
)

type Client interface {
	GetComponents(context.Context, uuid.UUID, *fleetdb.PaginationParams) (fleetdb.ServerComponentSlice, *fleetdb.ServerResponse, error)
	GetServerInventory(context.Context, uuid.UUID, bool) (*rivets.Server, *fleetdb.ServerResponse, error)
	UpdateServerInventory(context.Context, uuid.UUID, *rivets.Server, bool, *zap.Logger) error
	GetInventoryConverter() *inventoryconverter.InventoryConverter
}

// Creates a new Client, with reasonable defaults
func NewFleetDBClient(ctx context.Context, cfg *app.Configuration) (Client, error) {
	client, err := fleetdb.NewClient(cfg.FleetDBAddress, nil)
	if err != nil {
		return nil, err
	}

	if cfg.FleetDBToken != "" {
		client.SetToken(cfg.FleetDBToken)
	}

	// TODO: replace it with common.ComponentTypes() after figuring out
	// how to fetch ServerComponentType ID.
	// Then it's cleaner to move inventoryConverterInstance to the router
	// instead of the Client interface.
	slugs := make(map[string]*fleetdb.ServerComponentType)
	serverComponentTypes, _, err := client.ListServerComponentTypes(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get server component types: %w", err)
	}
	for _, ct := range serverComponentTypes {
		slugs[ct.Slug] = ct
	}

	return &fleetDBClient{
		client:                     client,
		inventoryConverterInstance: inventoryconverter.NewInventoryConverter(slugs),
	}, nil
}

type fleetDBClient struct {
	client                     *fleetdb.Client
	inventoryConverterInstance *inventoryconverter.InventoryConverter
}

func (fc fleetDBClient) GetServerInventory(ctx context.Context, id uuid.UUID, inband bool) (*rivets.Server, *fleetdb.ServerResponse, error) {
	return fc.client.GetServerInventory(ctx, id, inband)
}

func (fc fleetDBClient) GetComponents(ctx context.Context, id uuid.UUID, params *fleetdb.PaginationParams) (fleetdb.ServerComponentSlice, *fleetdb.ServerResponse, error) {
	return fc.client.GetComponents(ctx, id, params)
}

func (fc fleetDBClient) UpdateServerInventory(ctx context.Context, serverID uuid.UUID, rivetsServer *rivets.Server, inband bool, log *zap.Logger) error {
	if _, err := fc.client.SetServerInventory(ctx, serverID, rivetsServer, inband); err != nil {
		log.Error("set inventory fail", zap.String("server", serverID.String()), zap.String("err", err.Error()))
		return err
	}
	return nil
}

func (fc fleetDBClient) GetInventoryConverter() *inventoryconverter.InventoryConverter {
	return fc.inventoryConverterInstance
}
