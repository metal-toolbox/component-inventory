package internalfleetdb

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/metal-toolbox/component-inventory/internal/app"
	"github.com/metal-toolbox/component-inventory/pkg/api/constants"
	"github.com/metal-toolbox/component-inventory/pkg/api/types"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	"go.uber.org/zap"
)

type Client interface {
	GetServer(context.Context, uuid.UUID) (*fleetdb.Server, *fleetdb.ServerResponse, error)
	GetComponents(context.Context, uuid.UUID, *fleetdb.PaginationParams) (fleetdb.ServerComponentSlice, *fleetdb.ServerResponse, error)
	UpdateAttributes(context.Context, *fleetdb.Server, *types.ComponentInventoryDevice, *zap.Logger) error
	UpdateServerBIOSConfig(context.Context, *fleetdb.Server, *types.ComponentInventoryDevice, *zap.Logger) error
	UpdateServerMetadataAttributes(context.Context, uuid.UUID, *types.ComponentInventoryDevice, *zap.Logger) error
}

// Creates a new Client, with reasonable defaults
func NewFleetDBClient(cfg *app.Configuration) (Client, error) {
	client, err := fleetdb.NewClient(cfg.FleetDBAddress, nil)
	if err != nil {
		return nil, err
	}

	if cfg.FleetDBToken != "" {
		client.SetToken(cfg.FleetDBToken)
	}

	return &fleetDBClient{
		client: client,
	}, nil
}

type fleetDBClient struct {
	client  *fleetdb.Client
	appKind types.AppKind
}

func (fc fleetDBClient) GetServer(ctx context.Context, id uuid.UUID) (*fleetdb.Server, *fleetdb.ServerResponse, error) {
	return fc.client.Get(ctx, id)
}

func (fc fleetDBClient) GetComponents(ctx context.Context, id uuid.UUID, params *fleetdb.PaginationParams) (fleetdb.ServerComponentSlice, *fleetdb.ServerResponse, error) {
	return fc.client.GetComponents(ctx, id, params)
}

func (fc fleetDBClient) UpdateAttributes(ctx context.Context, server *fleetdb.Server, dev *types.ComponentInventoryDevice, log *zap.Logger) error {
	return createUpdateServerAttributes(ctx, fc.client, server, dev, log)
}

// Functions below may be refactored in the near future.
func createUpdateServerAttributes(ctx context.Context, c *fleetdb.Client, server *fleetdb.Server, dev *types.ComponentInventoryDevice, log *zap.Logger) error {
	newVendorData, newVendorAttrs, err := deviceVendorAttributes(dev)
	if err != nil {
		return err
	}

	// identify current vendor data in the inventory
	existingVendorAttrs := attributeByNamespace(constants.ServerVendorAttributeNS, server.Attributes)
	if existingVendorAttrs == nil {
		// create if none exists
		_, err = c.CreateAttributes(ctx, server.UUID, *newVendorAttrs)
		return err
	}

	// unpack vendor data from inventory
	existingVendorData := map[string]string{}
	if err := json.Unmarshal(existingVendorAttrs.Data, &existingVendorData); err != nil {
		// update vendor data since it seems to be invalid
		log.Warn("server vendor attributes data invalid, updating..")

		_, err = c.UpdateAttributes(ctx, server.UUID, constants.ServerVendorAttributeNS, newVendorAttrs.Data)

		return err
	}

	var changes bool
	for key := range newVendorData {
		if existingVendorData[key] == "" || existingVendorData[key] == "unknown" {
			if newVendorData[key] != "unknown" {
				changes = true
				existingVendorData[key] = newVendorData[key]
			}
		}
	}

	if !changes {
		return nil
	}

	if len(existingVendorData) > 0 {
		updateBytes, err := json.Marshal(existingVendorData)
		if err != nil {
			return err
		}

		_, err = c.UpdateAttributes(ctx, server.UUID, constants.ServerVendorAttributeNS, updateBytes)

		return err
	}

	return nil
}

func (fc fleetDBClient) UpdateServerBIOSConfig(ctx context.Context, server *fleetdb.Server, dev *types.ComponentInventoryDevice, log *zap.Logger) error {
	// Nothing to publish
	if len(*dev.BiosCfg) == 0 {
		log.Info("no bios configuration collected")
		return nil
	}

	// marshal metadata from device
	bc, err := json.Marshal(dev.BiosCfg)
	if err != nil {
		log.Error("invalid biosConfig")
		return err
	}

	va := fleetdb.VersionedAttributes{
		Namespace: serverBIOSConfigNS(fc.appKind),
		Data:      bc,
	}

	_, err = fc.client.CreateVersionedAttributes(ctx, server.UUID, va)

	return err
}

func (fc fleetDBClient) UpdateServerMetadataAttributes(ctx context.Context, serverID uuid.UUID, dev *types.ComponentInventoryDevice, log *zap.Logger) error {
	// no metadata reported in inventory from device
	if dev.Inv == nil || len(dev.Inv.Metadata) == 0 {
		// XXX: should delete the metadata on the server-service record!
		return nil
	}

	// marshal metadata from device
	metadata := mustFilterAssetMetadata(dev.Inv.Metadata)

	attribute := fleetdb.Attributes{
		Namespace: constants.ServerMetadataAttributeNS,
		Data:      metadata,
	}

	// XXX: This would be much easier if serverservice/fleetdb supported upsert
	// current asset metadata has no attributes set and no metadata attribute, create one
	if _, ok := dev.Inv.Metadata[ssMetadataAttributeFound]; !ok {
		_, err := fc.client.CreateAttributes(ctx, serverID, attribute)
		log.Info("creating server attributes")
		return err
	}

	// update vendor, model attributes
	_, err := fc.client.UpdateAttributes(ctx, serverID, constants.ServerMetadataAttributeNS, metadata)

	return err

}
