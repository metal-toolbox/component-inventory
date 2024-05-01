package inventoryconverter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/common"
	rivets "github.com/metal-toolbox/rivets/types"
)

type InventoryConverter struct {
	componentSlug map[string]bool
}

func NewInventoryConverter(componentSlug map[string]bool) *InventoryConverter {
	return &InventoryConverter{
		componentSlug: componentSlug,
	}
}

func (ic *InventoryConverter) ToRivetsServer(serverID, facility string, device *common.Device, biosCfg map[string]string) (*rivets.Server, error) {
	components, err := ic.getComponentSlice(device)
	if err != nil {
		return nil, err
	}

	deviceState := ""
	if device.Status != nil {
		deviceState = device.Status.State
	}

	return &rivets.Server{
		BIOSCfg:    biosCfg,
		ID:         serverID,
		Facility:   facility,
		Name:       serverID,
		Vendor:     device.Vendor,
		Model:      device.Model,
		Serial:     device.Serial,
		Status:     deviceState,
		Components: components,
	}, nil
}

func (ic *InventoryConverter) newComponent(slug, cvendor, cmodel, cserial, cproduct string, attrs *rivets.ComponentAttributes, status *common.Status, firmware *common.Firmware) *rivets.Component {
	slug = strings.ToLower(slug)
	_, exists := ic.componentSlug[slug]
	if !exists {
		return nil
	}

	if strings.TrimSpace(cmodel) == "" && strings.TrimSpace(cproduct) != "" {
		cmodel = cproduct
	}
	return &rivets.Component{
		Firmware:   firmware,
		Status:     status,
		Attributes: attrs,
		Name:       slug,
		Vendor:     cvendor,
		Model:      cmodel,
		Serial:     cserial,
	}
}

// nolint: gocyclo
func (ic *InventoryConverter) getComponentSlice(device *common.Device) ([]*rivets.Component, error) {
	if len(ic.componentSlug) == 0 {
		return nil, fmt.Errorf("omponent slugs lookup map empty")
	}
	components := []*rivets.Component{}

	// bios
	if devBIOS := device.BIOS; devBIOS != nil {
		if strings.TrimSpace(devBIOS.Serial) == "" {
			devBIOS.Serial = "0"
		}
		component := ic.newComponent(common.SlugBIOS, devBIOS.Vendor, devBIOS.Model, devBIOS.Serial, devBIOS.ProductName,
			&rivets.ComponentAttributes{
				Description:   devBIOS.Description,
				ProductName:   devBIOS.ProductName,
				SizeBytes:     devBIOS.SizeBytes,
				CapacityBytes: devBIOS.CapacityBytes,
				Oem:           devBIOS.Oem,
				Metadata:      devBIOS.Metadata,
				Capabilities:  devBIOS.Capabilities,
			},
			devBIOS.Status,
			devBIOS.Firmware,
		)
		if component != nil {
			components = append(components, component)
		}
	}

	// bmc
	if devBMC := device.BMC; devBMC != nil {
		if strings.TrimSpace(devBMC.Serial) == "" {
			devBMC.Serial = "0"
		}
		component := ic.newComponent(common.SlugBMC, devBMC.Vendor, devBMC.Model, devBMC.Serial, devBMC.ProductName,
			&rivets.ComponentAttributes{
				Description:  devBMC.Description,
				ProductName:  devBMC.ProductName,
				Oem:          devBMC.Oem,
				Metadata:     devBMC.Metadata,
				Capabilities: devBMC.Capabilities,
			},
			devBMC.Status,
			devBMC.Firmware,
		)
		if component != nil {
			components = append(components, component)
		}
	}

	// mainboard
	if devMainBoard := device.Mainboard; devMainBoard != nil {
		if strings.TrimSpace(devMainBoard.Serial) == "" {
			devMainBoard.Serial = "0"
		}
		component := ic.newComponent(common.SlugMainboard, devMainBoard.Vendor, devMainBoard.Model, devMainBoard.Serial, devMainBoard.ProductName,
			&rivets.ComponentAttributes{
				Description:  devMainBoard.Description,
				ProductName:  devMainBoard.ProductName,
				Oem:          devMainBoard.Oem,
				PhysicalID:   devMainBoard.PhysicalID,
				Metadata:     devMainBoard.Metadata,
				Capabilities: devMainBoard.Capabilities,
			},
			devMainBoard.Status,
			devMainBoard.Firmware,
		)
		if component != nil {
			components = append(components, component)
		}
	}

	// memory
	if devMemorys := device.Memory; len(devMemorys) > 0 {
		for idx, dm := range devMemorys {
			if dm.Vendor == "" && dm.ProductName == "" && dm.SizeBytes == 0 && dm.ClockSpeedHz == 0 {
				continue
			}
			// set incrementing serial when one isn't found
			if strings.TrimSpace(dm.Serial) == "" {
				dm.Serial = strconv.Itoa(idx)
			}
			// trim redundant prefix
			dm.Slot = strings.TrimPrefix(dm.Slot, "DIMM.Socket.")
			component := ic.newComponent(common.SlugPhysicalMem, dm.Vendor, dm.Model, dm.Serial, dm.ProductName,
				&rivets.ComponentAttributes{
					Description:  dm.Description,
					ProductName:  dm.ProductName,
					Oem:          dm.Oem,
					Slot:         dm.Slot,
					ClockSpeedHz: dm.ClockSpeedHz,
					FormFactor:   dm.FormFactor,
					PartNumber:   dm.PartNumber,
					Metadata:     dm.Metadata,
					SizeBytes:    dm.SizeBytes,
					Capabilities: dm.Capabilities,
				},
				dm.Status,
				dm.Firmware,
			)
			if component != nil {
				components = append(components, component)
			}
		}
	}

	// nics
	if devNICs := device.NICs; len(devNICs) > 0 {
		for idx, dn := range devNICs {
			if strings.TrimSpace(dn.Serial) == "" {
				dn.Serial = strconv.Itoa(idx)
			}
			component := ic.newComponent(common.SlugNIC, dn.Vendor, dn.Model, dn.Serial, dn.ProductName, nil, dn.Status, dn.Firmware)
			if component != nil {
				components = append(components, component)
			}
		}
	}

	// drives
	if devDrives := device.Drives; len(devDrives) > 0 {
		for idx, dd := range devDrives {
			if strings.TrimSpace(dd.Serial) == "" {
				dd.Serial = strconv.Itoa(idx)
			}
			component := ic.newComponent(common.SlugDrive, dd.Vendor, dd.Model, dd.Serial, dd.ProductName,
				&rivets.ComponentAttributes{
					Description:         dd.Description,
					ProductName:         dd.ProductName,
					Oem:                 dd.Oem,
					Metadata:            dd.Metadata,
					BusInfo:             dd.BusInfo,
					OemID:               dd.OemID,
					StorageController:   dd.StorageController,
					Protocol:            dd.Protocol,
					SmartErrors:         dd.SmartErrors,
					SmartStatus:         dd.SmartStatus,
					DriveType:           dd.Type,
					WWN:                 dd.WWN,
					CapacityBytes:       dd.CapacityBytes,
					BlockSizeBytes:      dd.BlockSizeBytes,
					CapableSpeedGbps:    dd.CapableSpeedGbps,
					NegotiatedSpeedGbps: dd.NegotiatedSpeedGbps,
					Capabilities:        dd.Capabilities,
				},
				dd.Status,
				dd.Firmware,
			)
			if component != nil {
				// some drives show up with model numbers in the description field.
				if component.Model == "" && dd.Description != "" {
					component.Model = dd.Description
				}
				components = append(components, component)
			}
		}
	}

	// psus
	if devPSUs := device.PSUs; len(devPSUs) > 0 {
		for idx, dp := range devPSUs {
			if strings.TrimSpace(dp.Serial) == "" {
				dp.Serial = strconv.Itoa(idx)
			}
			component := ic.newComponent(common.SlugPSU, dp.Vendor, dp.Model, dp.Serial, dp.ProductName,
				&rivets.ComponentAttributes{
					ID:                 dp.ID,
					Description:        dp.Description,
					ProductName:        dp.ProductName,
					PowerCapacityWatts: dp.PowerCapacityWatts,
					Oem:                dp.Oem,
					Metadata:           dp.Metadata,
					Capabilities:       dp.Capabilities,
				},
				dp.Status,
				dp.Firmware,
			)
			if component != nil {
				components = append(components, component)
			}
		}
	}

	// cpus
	if devCPUs := device.CPUs; len(devCPUs) > 0 {
		for idx, dc := range devCPUs {
			if strings.TrimSpace(dc.Serial) == "" {
				dc.Serial = strconv.Itoa(idx)
			}
			component := ic.newComponent(common.SlugCPU, dc.Vendor, dc.Model, dc.Serial, dc.ProductName,
				&rivets.ComponentAttributes{
					ID:           dc.ID,
					Description:  dc.Description,
					ProductName:  dc.ProductName,
					Metadata:     dc.Metadata,
					Slot:         dc.Slot,
					Architecture: dc.Architecture,
					ClockSpeedHz: dc.ClockSpeedHz,
					Cores:        dc.Cores,
					Threads:      dc.Threads,
					Capabilities: dc.Capabilities,
				},
				dc.Status,
				dc.Firmware,
			)
			if component != nil {
				components = append(components, component)
			}
		}
	}

	// tpms
	if devTPMs := device.TPMs; len(devTPMs) > 0 {
		for idx, dt := range devTPMs {
			if strings.TrimSpace(dt.Serial) == "" {
				dt.Serial = strconv.Itoa(idx)
			}
			component := ic.newComponent(common.SlugTPM, dt.Vendor, dt.Model, dt.Serial, dt.ProductName,
				&rivets.ComponentAttributes{
					Description:   dt.Description,
					ProductName:   dt.ProductName,
					Metadata:      dt.Metadata,
					Capabilities:  dt.Capabilities,
					InterfaceType: dt.InterfaceType,
				},
				dt.Status,
				dt.Firmware,
			)
			if component != nil {
				components = append(components, component)
			}
		}
	}

	// cplds
	if devCPLDs := device.CPLDs; len(devCPLDs) > 0 {
		for idx, dc := range devCPLDs {
			if strings.TrimSpace(dc.Serial) == "" {
				dc.Serial = strconv.Itoa(idx)
			}
			component := ic.newComponent(common.SlugCPLD, dc.Vendor, dc.Model, dc.Serial, dc.ProductName,
				&rivets.ComponentAttributes{
					Description:  dc.Description,
					ProductName:  dc.ProductName,
					Metadata:     dc.Metadata,
					Capabilities: dc.Capabilities,
				},
				dc.Status,
				dc.Firmware,
			)
			if component != nil {
				components = append(components, component)
			}
		}
	}

	// gpus
	if devGPUs := device.GPUs; len(devGPUs) > 0 {
		for idx, dg := range devGPUs {
			if strings.TrimSpace(dg.Serial) == "" {
				dg.Serial = strconv.Itoa(idx)
			}
			component := ic.newComponent(common.SlugGPU, dg.Vendor, dg.Model, dg.Serial, dg.ProductName,
				&rivets.ComponentAttributes{
					Description:  dg.Description,
					ProductName:  dg.ProductName,
					Metadata:     dg.Metadata,
					Capabilities: dg.Capabilities,
				},
				dg.Status,
				dg.Firmware,
			)
			if component != nil {
				components = append(components, component)
			}
		}
	}

	// storage controllers
	if devStorages := device.StorageControllers; len(devStorages) > 0 {
		for idx, ds := range devStorages {
			if strings.TrimSpace(ds.Serial) == "" {
				ds.Serial = strconv.Itoa(idx)
			}
			component := ic.newComponent(common.SlugStorageController, ds.Vendor, ds.Model, ds.Serial, ds.ProductName,
				&rivets.ComponentAttributes{
					ID:                           ds.ID,
					Description:                  ds.Description,
					ProductName:                  ds.ProductName,
					Oem:                          ds.Oem,
					SupportedControllerProtocols: ds.SupportedControllerProtocols,
					SupportedDeviceProtocols:     ds.SupportedDeviceProtocols,
					SupportedRAIDTypes:           ds.SupportedRAIDTypes,
					PhysicalID:                   ds.PhysicalID,
					BusInfo:                      ds.BusInfo,
					SpeedGbps:                    ds.SpeedGbps,
					Metadata:                     ds.Metadata,
					Capabilities:                 ds.Capabilities,
				},
				ds.Status,
				ds.Firmware,
			)
			if component != nil {
				// some controller show up with model numbers in the description field.
				if component.Model == "" && ds.Description != "" {
					component.Model = ds.Description
				}
				components = append(components, component)
			}
		}
	}

	// enclosures
	if devEnclosures := device.Enclosures; len(devEnclosures) > 0 {
		for idx, de := range devEnclosures {
			if strings.TrimSpace(de.Serial) == "" {
				de.Serial = strconv.Itoa(idx)
			}
			component := ic.newComponent(common.SlugEnclosure, de.Vendor, de.Model, de.Serial, de.ProductName,
				&rivets.ComponentAttributes{
					ID:           de.ID,
					Description:  de.Description,
					ProductName:  de.ProductName,
					Oem:          de.Oem,
					Metadata:     de.Metadata,
					ChassisType:  de.ChassisType,
					Capabilities: de.Capabilities,
				},
				de.Status,
				de.Firmware,
			)
			if component != nil {
				components = append(components, component)
			}
		}
	}

	return components, nil
}
