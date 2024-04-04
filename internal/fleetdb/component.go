package internalfleetdb

import (
	"github.com/bmc-toolbox/common"
	"github.com/google/uuid"

	fleetdbapi "github.com/metal-toolbox/fleetdb/pkg/api/v1"
)

// attributes are generic component attributes
type attributes struct {
	Capabilities                 []*common.Capability `json:"capabilities,omitempty"`
	Metadata                     map[string]string    `json:"metadata,omitempty"`
	ID                           string               `json:"id,omitempty"`
	ChassisType                  string               `json:"chassis_type,omitempty"`
	Description                  string               `json:"description,omitempty"`
	ProductName                  string               `json:"product_name,omitempty"`
	InterfaceType                string               `json:"interface_type,omitempty"`
	Slot                         string               `json:"slot,omitempty"`
	Architecture                 string               `json:"architecture,omitempty"`
	MacAddress                   string               `json:"macaddress,omitempty"`
	SupportedControllerProtocols string               `json:"supported_controller_protocol,omitempty"`
	SupportedDeviceProtocols     string               `json:"supported_device_protocol,omitempty"`
	SupportedRAIDTypes           string               `json:"supported_raid_types,omitempty"`
	PhysicalID                   string               `json:"physid,omitempty"`
	FormFactor                   string               `json:"form_factor,omitempty"`
	PartNumber                   string               `json:"part_number,omitempty"`
	OemID                        string               `json:"oem_id,omitempty"`
	DriveType                    string               `json:"drive_type,omitempty"`
	StorageController            string               `json:"storage_controller,omitempty"`
	BusInfo                      string               `json:"bus_info,omitempty"`
	WWN                          string               `json:"wwn,omitempty"`
	Protocol                     string               `json:"protocol,omitempty"`
	SmartStatus                  string               `json:"smart_status,omitempty"`
	SmartErrors                  []string             `json:"smart_errors,omitempty"`
	PowerCapacityWatts           int64                `json:"power_capacity_watts,omitempty"`
	SizeBytes                    int64                `json:"size_bytes,omitempty"`
	CapacityBytes                int64                `json:"capacity_bytes,omitempty" diff:"immutable"`
	ClockSpeedHz                 int64                `json:"clock_speed_hz,omitempty"`
	Cores                        int                  `json:"cores,omitempty"`
	Threads                      int                  `json:"threads,omitempty"`
	SpeedBits                    int64                `json:"speed_bits,omitempty"`
	SpeedGbps                    int64                `json:"speed_gbps,omitempty"`
	BlockSizeBytes               int64                `json:"block_size_bytes,omitempty"`
	CapableSpeedGbps             int64                `json:"capable_speed_gbps,omitempty"`
	NegotiatedSpeedGbps          int64                `json:"negotiated_speed_gbps,omitempty"`
	Oem                          bool                 `json:"oem,omitempty"`
}

// toComponentSlice converts a common.Device object to the fleetdb component slice object
func toComponentSlice(serverID uuid.UUID, inventory *common.Device) ([]*fleetdbapi.ServerComponent, error) {
	componentsTmp := []*fleetdbapi.ServerComponent{}

	// vendor := inventory.Vendor
	// componentsTmp = append(componentsTmp,
	// 	bios(vendor, inventory.BIOS),
	// 	bmc(vendor, inventory.BMC),
	// 	mainboard(vendor, inventory.Mainboard),
	// )

	// componentsTmp = append(componentsTmp, dimms(vendor, inventory.Memory)...)
	// componentsTmp = append(componentsTmp, nics(vendor, inventory.NICs)...)
	// componentsTmp = append(componentsTmp, drives(vendor, inventory.Drives)...)
	// componentsTmp = append(componentsTmp, psus(vendor, inventory.PSUs)...)
	// componentsTmp = append(componentsTmp, cpus(vendor, inventory.CPUs)...)
	// componentsTmp = append(componentsTmp, tpms(vendor, inventory.TPMs)...)
	// componentsTmp = append(componentsTmp, cplds(vendor, inventory.CPLDs)...)
	// componentsTmp = append(componentsTmp, gpus(vendor, inventory.GPUs)...)
	// componentsTmp = append(componentsTmp, storageControllers(vendor, inventory.StorageControllers)...)
	// componentsTmp = append(componentsTmp, enclosures(vendor, inventory.Enclosures)...)

	components := []*fleetdbapi.ServerComponent{}

	for _, component := range componentsTmp {
		if component == nil || requiredAttributesEmpty(component) {
			continue
		}

		component.ServerUUID = serverID
		components = append(components, component)
	}

	return components, nil
}

func requiredAttributesEmpty(component *fleetdbapi.ServerComponent) bool {
	return component.Serial == "0" &&
		component.Model == "" &&
		component.Vendor == "" &&
		len(component.Attributes) == 0 &&
		len(component.VersionedAttributes) == 0
}
