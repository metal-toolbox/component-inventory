package routes

import (
	"context"
	"errors"

	internalfleetdb "github.com/metal-toolbox/component-inventory/internal/fleetdb"
	"github.com/metal-toolbox/component-inventory/pkg/api/types"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	"go.uber.org/zap"
)

func processInband(ctx context.Context, c internalfleetdb.Client, server *fleetdb.Server, dev *types.ComponentInventoryDevice, log *zap.Logger) error { //nolint
	log.Info("processing", zap.String("server", server.Name), zap.String("device", dev.Inv.Serial))
	// update bisconfig
	if err := c.UpdateServerBIOSConfig(ctx, server, dev, log); err != nil {
		return err
	}

	// create/update server serial, vendor, model attributes
	if err := c.UpdateAttributes(ctx, server, dev, log); err != nil {
		return err
	}

	// create update server metadata attributes
	if err := c.UpdateServerMetadataAttributes(ctx, server.UUID, dev, log); err != nil {
		return err
	}

	return errors.New("not implemented")
}

func processOutofband(ctx context.Context, c internalfleetdb.Client, server *fleetdb.Server, dev *types.ComponentInventoryDevice, log *zap.Logger) error { //nolint
	log.Info("processing", zap.String("server", server.Name), zap.String("device", dev.Inv.Serial))
	return errors.New("not implemented")
}
