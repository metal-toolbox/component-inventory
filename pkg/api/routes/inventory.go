package routes

import (
	"context"

	"github.com/google/uuid"
	"github.com/metal-toolbox/alloy/types"
	internalfleetdb "github.com/metal-toolbox/component-inventory/internal/fleetdb"
	rivets "github.com/metal-toolbox/rivets/types"
	"go.uber.org/zap"
)

func processInband(ctx context.Context, c internalfleetdb.Client, server *rivets.Server, dev *types.InventoryDevice, log *zap.Logger) error { //nolint
	log.Info("processing", zap.String("server", server.Name), zap.String("device", dev.Inv.Serial))
	srvID, err := uuid.Parse(server.ID)
	if err != nil {
		log.Error("failed to parse server ID", zap.String("server", server.ID), zap.String("err", err.Error()))
		return err
	}

	rivetsServer, err := c.GetInventoryConverter().ToRivetsServer(server.ID, server.Facility, dev.Inv, dev.BiosCfg)
	if err != nil {
		log.Error("convert inventory fail", zap.String("server", server.Name), zap.String("err", err.Error()))
		return err
	}

	compareComponents(server, rivetsServer, log)

	return c.UpdateServerInventory(ctx, srvID, rivetsServer, true, log)
}

func processOutofband(ctx context.Context, c internalfleetdb.Client, server *rivets.Server, dev *types.InventoryDevice, log *zap.Logger) error { //nolint
	log.Info("processing", zap.String("server", server.ID), zap.String("device", dev.Inv.Serial))
	srvID, err := uuid.Parse(server.ID)
	if err != nil {
		log.Error("failed to parse server ID", zap.String("server", server.ID), zap.String("err", err.Error()))
		return err
	}

	rivetsServer, err := c.GetInventoryConverter().ToRivetsServer(server.ID, server.Facility, dev.Inv, dev.BiosCfg)
	if err != nil {
		log.Error("convert inventory fail", zap.String("server", server.Name), zap.String("err", err.Error()))
		return err
	}

	compareComponents(server, rivetsServer, log)

	return c.UpdateServerInventory(ctx, srvID, rivetsServer, false, log)
}
