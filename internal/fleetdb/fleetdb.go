package internalfleetdb

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/metal-toolbox/alloy/types"
	"github.com/metal-toolbox/component-inventory/internal/app"
	"github.com/metal-toolbox/component-inventory/internal/inventoryconverter"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	"go.uber.org/zap"
)

type Client interface {
	GetServer(context.Context, uuid.UUID) (*fleetdb.Server, *fleetdb.ServerResponse, error)
	GetComponents(context.Context, uuid.UUID, *fleetdb.PaginationParams) (fleetdb.ServerComponentSlice, *fleetdb.ServerResponse, error)
	UpdateServerInventory(context.Context, *fleetdb.Server, *types.InventoryDevice, *zap.Logger, bool) error
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
		slugs:                      slugs,
		inventoryConverterInstance: inventoryconverter.NewInventoryConverter(slugs),
	}, nil
}

type fleetDBClient struct {
	client                     *fleetdb.Client
	slugs                      map[string]*fleetdb.ServerComponentType
	inventoryConverterInstance *inventoryconverter.InventoryConverter
}

func (fc fleetDBClient) GetServer(ctx context.Context, id uuid.UUID) (*fleetdb.Server, *fleetdb.ServerResponse, error) {
	return fc.client.Get(ctx, id)
}

func (fc fleetDBClient) GetComponents(ctx context.Context, id uuid.UUID, params *fleetdb.PaginationParams) (fleetdb.ServerComponentSlice, *fleetdb.ServerResponse, error) {
	return fc.client.GetComponents(ctx, id, params)
}

func (fc fleetDBClient) UpdateServerInventory(ctx context.Context, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger, inband bool) error {
	log.Info("update server inventory", zap.String("server", server.Name))
	rivetsServer, err := fc.inventoryConverterInstance.ToRivetsServer(server.UUID.String(), server.FacilityCode, dev.Inv, dev.BiosCfg)
	if err != nil {
		log.Error("convert inventory fail", zap.String("server", server.Name), zap.String("err", err.Error()))
		return err
	}
	_, err = fc.client.SetServerInventory(ctx, server.UUID, rivetsServer, inband)
	if err != nil {
		log.Error("set inventory fail", zap.String("server", server.Name), zap.String("err", err.Error()))
		return err
	}
	return nil
}
