package routes

import (
	"context"
	"errors"

	"github.com/metal-toolbox/alloy/types"
	internalfleetdb "github.com/metal-toolbox/component-inventory/internal/fleetdb"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	"go.uber.org/zap"
)

func processInband(ctx context.Context, c internalfleetdb.Client, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger) error { //nolint
	log.Info("processing", zap.String("server", server.Name), zap.String("device", dev.Inv.Serial))
	if err := c.UpdateAttributes(ctx, server, dev, log); err != nil {
		return err
	}
	return errors.New("not implemented")
}

func processOutofband(ctx context.Context, c internalfleetdb.Client, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger) error { //nolint
	log.Info("processing", zap.String("server", server.Name), zap.String("device", dev.Inv.Serial))
	return errors.New("not implemented")
}
