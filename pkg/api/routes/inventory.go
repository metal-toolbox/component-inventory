package routes

import (
	"context"
	"fmt"
	"strings"

	"github.com/metal-toolbox/alloy/types"
	internalfleetdb "github.com/metal-toolbox/component-inventory/internal/fleetdb"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	"go.uber.org/zap"
)

func processInband(ctx context.Context, c internalfleetdb.Client, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger) error { //nolint
	log.Info("processing", zap.String("server", server.Name), zap.String("device", dev.Inv.Serial))
	if err := verifyComponent(c, server, dev, log); err != nil {
		return err
	}
	return c.UpdateServerInventory(ctx, server, dev, log, true)
}

func processOutofband(ctx context.Context, c internalfleetdb.Client, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger) error { //nolint
	log.Info("processing", zap.String("server", server.Name), zap.String("device", dev.Inv.Serial))
	if err := verifyComponent(c, server, dev, log); err != nil {
		return err
	}
	return c.UpdateServerInventory(ctx, server, dev, log, false)
}

func verifyComponent(c internalfleetdb.Client, server *fleetdb.Server, dev *types.InventoryDevice, log *zap.Logger) error {
	components, err := fetchServerComponents(c, server.UUID, log)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			// The server doesn't have components, we can create it.
			return nil
		}
		return err
	}
	return isBadComponents(components, dev)
}

func isBadComponents(_ serverComponents, _ *types.InventoryDevice) error {
	return fmt.Errorf("unimplement")
}
