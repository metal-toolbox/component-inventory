package types

import (
	"github.com/bmc-toolbox/common"
)

type (
	BiosConfig map[string]string
	AppKind    string
)

type ComponentInventoryDevice struct {
	ID      string         `json:"id,omitempty"`
	Inv     *common.Device `json:"inventory,omitempty"`
	BiosCfg *BiosConfig    `json:"biosconfig,omitempty"`
}
