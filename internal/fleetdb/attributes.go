package internalfleetdb

import (
	"encoding/json"

	"github.com/metal-toolbox/alloy/types"
	"github.com/metal-toolbox/component-inventory/pkg/api/constants"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
)

const (
	uefiVariablesKey         = "uefi-variables"
	ssMetadataAttributeFound = "__ss_found"
)

func deviceVendorAttributes(cid *types.InventoryDevice) (map[string]string, *fleetdb.Attributes, error) {
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

// mustFilterInventoryDeviceMetadata processes the asset inventory metadata to filter out fields we'll turn into versioned attributes (e.g. UEFIVariables)
func mustFilterInventoryDeviceMetadata(inventory map[string]string) json.RawMessage {
	excludedKeys := map[string]struct{}{
		uefiVariablesKey: {},
	}

	filtered := make(map[string]string)

	for k, v := range inventory {
		if _, ok := excludedKeys[k]; ok {
			continue
		}
		filtered[k] = v
	}

	byt, err := json.Marshal(filtered)
	if err != nil {
		panic("serializing metadata string map")
	}

	return byt
}
