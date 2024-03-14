package routes

import (
	"errors"

	"github.com/bmc-toolbox/common"
	"github.com/google/uuid"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	"go.uber.org/zap"
)

func processInband(c *fleetdb.Client, srvID uuid.UUID, dev *common.Device, log *zap.Logger) error { //nolint
	log.Info("processing", zap.String("server id", srvID.String()), zap.String("device", dev.Serial))
	return errors.New("not implemented")
}

func processOutofband(c *fleetdb.Client, srvID uuid.UUID, dev *common.Device, log *zap.Logger) error { //nolint
	log.Info("processing", zap.String("server id", srvID.String()), zap.String("device", dev.Serial))
	return errors.New("not implemented")
}
