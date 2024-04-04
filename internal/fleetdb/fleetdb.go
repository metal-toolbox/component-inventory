package internalfleetdb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/metal-toolbox/alloy/types"
	"github.com/metal-toolbox/component-inventory/internal/app"
	"github.com/metal-toolbox/component-inventory/pkg/api/constants"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type Client interface {
	GetServer(context.Context, uuid.UUID) (*fleetdb.Server, *fleetdb.ServerResponse, error)
	GetComponents(context.Context, uuid.UUID, *fleetdb.PaginationParams) (fleetdb.ServerComponentSlice, *fleetdb.ServerResponse, error)
	UpdateInventory(context.Context, *fleetdb.Server, *types.InventoryDevice, *zap.Logger) error
	UpdateBIOSConfigration(context.Context, *fleetdb.Server, *types.InventoryDevice, *zap.Logger) error
}

type AppKind string

const (
	AppKindInband    AppKind = "inband"
	AppKindOutOfBand AppKind = "outofband"
	pkgName                  = "internal/store"
)

var ErrInventoryQuery = errors.New("inventory query returned error")

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
	client *fleetdb.Client
}

func (fc fleetDBClient) GetServer(ctx context.Context, id uuid.UUID) (*fleetdb.Server, *fleetdb.ServerResponse, error) {
	return fc.client.Get(ctx, id)
}

func (fc fleetDBClient) GetComponents(ctx context.Context, id uuid.UUID, params *fleetdb.PaginationParams) (fleetdb.ServerComponentSlice, *fleetdb.ServerResponse, error) {
	return fc.client.GetComponents(ctx, id, params)
}

func (fc fleetDBClient) UpdateInventory(ctx context.Context, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger) error {
	// create/update server serial, vendor, model attributes
	if err := fc.CreateUpdateServerAttributes(ctx, server, dev, log); err != nil {
		return errors.Wrap(ErrInventoryQuery, "Server Vendor attribute create/update error: "+err.Error())
	}
	// create update server metadata attributes
	if err := fc.CreateUpdateServerMetadataAttributes(ctx, server, dev, log); err != nil {
		return errors.Wrap(ErrInventoryQuery, "Server Metadata attribute create/update error: "+err.Error())
	}
	// create update server component
	if err := fc.CreateUpdateServerComponents(ctx, server, dev, log); err != nil {
		return errors.Wrap(ErrInventoryQuery, "Server Component create/update error: "+err.Error())
	}
	return nil
}

func (fc fleetDBClient) CreateUpdateServerAttributes(ctx context.Context, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger) error {
	newVendorData, newVendorAttrs, err := deviceVendorAttributes(dev)
	if err != nil {
		return err
	}

	// identify current vendor data in the inventory
	existingVendorAttrs := attributeByNamespace(constants.ServerVendorAttributeNS, server.Attributes)
	if existingVendorAttrs == nil {
		// create if none exists
		_, err = fc.client.CreateAttributes(ctx, server.UUID, *newVendorAttrs)
		return err
	}

	// unpack vendor data from inventory
	existingVendorData := map[string]string{}
	if err := json.Unmarshal(existingVendorAttrs.Data, &existingVendorData); err != nil {
		// update vendor data since it seems to be invalid
		log.Warn("server vendor attributes data invalid, updating..")

		_, err = fc.client.UpdateAttributes(ctx, server.UUID, constants.ServerVendorAttributeNS, newVendorAttrs.Data)

		return err
	}

	updatedVendorData := existingVendorData
	var changes bool
	for key := range newVendorData {
		if updatedVendorData[key] == "" || updatedVendorData[key] == "unknown" {
			if newVendorData[key] != "unknown" {
				changes = true
				updatedVendorData[key] = newVendorData[key]
			}
		}
	}

	if !changes {
		return nil
	}

	if len(updatedVendorData) > 0 {
		updateBytes, err := json.Marshal(updatedVendorData)
		if err != nil {
			return err
		}

		_, err = fc.client.UpdateAttributes(ctx, server.UUID, constants.ServerVendorAttributeNS, updateBytes)

		return err
	}

	return nil
}

func (fc fleetDBClient) CreateUpdateServerMetadataAttributes(ctx context.Context, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger) error {
	// no metadata reported in inventory from device
	if dev.Inv == nil || len(dev.Inv.Metadata) == 0 {
		// XXX: should delete the metadata on the server-service record!
		return nil
	}

	// marshal metadata from device
	metadata := mustFilterInventoryDeviceMetadata(dev.Inv.Metadata)

	attribute := fleetdb.Attributes{
		Namespace: constants.ServerMetadataAttributeNS,
		Data:      metadata,
	}

	// XXX: This would be much easier if serverservice/fleetdb supported upsert
	// current asset metadata has no attributes set and no metadata attribute, create one
	serverID := server.UUID
	if _, ok := dev.Inv.Metadata[ssMetadataAttributeFound]; !ok {
		_, err := fc.client.CreateAttributes(ctx, serverID, attribute)
		log.Info("creating server attributes")
		return err
	}

	// update vendor, model attributes
	_, err := fc.client.UpdateAttributes(ctx, serverID, constants.ServerMetadataAttributeNS, metadata)

	return err

}

// CreateUpdateServerComponents compares the current object in fleetdb with the device data and creates/updates server component data.
//
// nolint:gocyclo // the method caries out all steps to have device data compared and registered, for now its accepted as cyclomatic.
func (fc fleetDBClient) CreateUpdateServerComponents(ctx context.Context, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger) error {
	ctx, span := otel.Tracer(pkgName).Start(ctx, "Serverservice.createUpdateServerComponents")
	defer span.End()

	if dev == nil {
		return nil
	}

	serverID := server.UUID
	// convert model.AssetDevice to server service component slice
	_, err := toComponentSlice(serverID, dev.Inv)
	if err != nil {
		return errors.Wrap(ErrInventoryDeviceObjectConversion, err.Error())
	}

	return fmt.Errorf("unimplemented")
}

func (fc fleetDBClient) UpdateBIOSConfigration(ctx context.Context, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger) error {
	// marshal metadata from device
	bc, err := json.Marshal(dev.BiosCfg)
	if err != nil {
		return err
	}

	biosConfigNS := os.Getenv("ALLOY_FLEETDB_BIOS_CONFIG_NS")
	if biosConfigNS == "" {
		biosConfigNS = fmt.Sprintf("%s.bios_configuration", constants.FleetDBNSPrefix)
	}

	va := fleetdb.VersionedAttributes{
		Namespace: biosConfigNS,
		Data:      bc,
	}

	_, err = fc.client.CreateVersionedAttributes(ctx, server.UUID, va)

	return err
}
