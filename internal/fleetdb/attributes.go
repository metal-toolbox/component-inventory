package internalfleetdb

import (
	"encoding/json"

	"github.com/metal-toolbox/component-inventory/pkg/api/constants"
	"github.com/metal-toolbox/component-inventory/pkg/api/types"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
)

func deviceVendorAttributes(cid *types.ComponentInventoryDevice) (map[string]string, *fleetdb.Attributes, error) {
	deviceVendorData := map[string]string{
		constants.ServerSerialAttributeKey: "unknown",
		constants.ServerVendorAttributeKey: "unknown",
		constants.ServerModelAttributeKey:  "unknown",
	}

	if cid.Inv != nil {
		if cid.Inv.Serial != "" {
			deviceVendorData[constants.ServerSerialAttributeKey] = cid.Inv.Serial
		}

		if cid.Inv.Model != "" {
			deviceVendorData[constants.ServerModelAttributeKey] = cid.Inv.Model
		}

		if cid.Inv.Vendor != "" {
			deviceVendorData[constants.ServerVendorAttributeKey] = cid.Inv.Vendor
		}
	}

	deviceVendorDataBytes, err := json.Marshal(deviceVendorData)
	if err != nil {
		return nil, nil, err
	}

	return deviceVendorData, &fleetdb.Attributes{
		Namespace: constants.ServerVendorAttributeNS,
		Data:      deviceVendorDataBytes,
	}, nil
}

// attributeByNamespace returns the attribute in the slice that matches the namespace
func attributeByNamespace(ns string, attributes []fleetdb.Attributes) *fleetdb.Attributes {
	for _, attribute := range attributes {
		if attribute.Namespace == ns {
			return &attribute
		}
	}

	return nil
}
