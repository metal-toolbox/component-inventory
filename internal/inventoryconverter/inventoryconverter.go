package inventoryconverter

import (
	"strconv"
	"strings"

	"github.com/bmc-toolbox/common"
	rivets "github.com/metal-toolbox/rivets/types"
)

func ToRivetsServer(serverID, facility string, device *common.Device, biosCfg map[string]string) *rivets.Server {
	components := getComponentSlice(device)

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
	}
}

func newComponent(slug, cvendor, cmodel, cserial, cproduct string, attrs *rivets.ComponentAttributes, status *common.Status, firmware *common.Firmware) *rivets.Component {
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

func biosToComponent(b *common.BIOS) *rivets.Component {
	serial := b.Serial
	if strings.TrimSpace(serial) == "" {
		serial = "0"
	}
	component := newComponent(common.SlugBIOS, b.Vendor, b.Model, serial, b.ProductName,
		&rivets.ComponentAttributes{
			Description:   b.Description,
			ProductName:   b.ProductName,
			SizeBytes:     b.SizeBytes,
			CapacityBytes: b.CapacityBytes,
			Oem:           b.Oem,
			Metadata:      b.Metadata,
			Capabilities:  b.Capabilities,
		},
		b.Status,
		b.Firmware,
	)
	return component
}

func bmcToComponent(b *common.BMC) *rivets.Component {
	serial := b.Serial
	if strings.TrimSpace(serial) == "" {
		serial = "0"
	}
	component := newComponent(common.SlugBMC, b.Vendor, b.Model, serial, b.ProductName,
		&rivets.ComponentAttributes{
			Description:  b.Description,
			ProductName:  b.ProductName,
			Oem:          b.Oem,
			Metadata:     b.Metadata,
			Capabilities: b.Capabilities,
		},
		b.Status,
		b.Firmware,
	)
	return component
}

func mainboardToComponent(m *common.Mainboard) *rivets.Component {
	serial := m.Serial
	if strings.TrimSpace(serial) == "" {
		serial = "0"
	}
	component := newComponent(
		common.SlugMainboard, m.Vendor,
		m.Model, serial, m.ProductName,
		&rivets.ComponentAttributes{
			Description:  m.Description,
			ProductName:  m.ProductName,
			Oem:          m.Oem,
			PhysicalID:   m.PhysicalID,
			Metadata:     m.Metadata,
			Capabilities: m.Capabilities,
		},
		m.Status,
		m.Firmware,
	)
	return component
}

func dimmsToComponentSlice(ds []*common.Memory) []*rivets.Component {
	comps := []*rivets.Component{}
	for idx, d := range ds {
		serial := d.Serial
		// set incrementing serial when one isn't found
		if strings.TrimSpace(serial) == "" {
			serial = strconv.Itoa(idx)
		}
		// trim redundant prefix
		slot := strings.TrimPrefix(d.Slot, "DIMM.Socket.")
		component := newComponent(common.SlugPhysicalMem, d.Vendor, d.Model, serial, d.ProductName,
			&rivets.ComponentAttributes{
				Description:  d.Description,
				ProductName:  d.ProductName,
				Oem:          d.Oem,
				Slot:         slot,
				ClockSpeedHz: d.ClockSpeedHz,
				FormFactor:   d.FormFactor,
				PartNumber:   d.PartNumber,
				Metadata:     d.Metadata,
				SizeBytes:    d.SizeBytes,
				Capabilities: d.Capabilities,
			},
			d.Status,
			d.Firmware,
		)
		comps = append(comps, component)
	}
	return comps
}

func nicsToComponentSlice(ns []*common.NIC) []*rivets.Component {
	comps := []*rivets.Component{}
	for idx, n := range ns {
		serial := n.Serial
		if strings.TrimSpace(serial) == "" {
			serial = strconv.Itoa(idx)
		}
		component := newComponent(common.SlugNIC, n.Vendor, n.Model, serial, n.ProductName, nil, n.Status, n.Firmware)
		comps = append(comps, component)
	}
	return comps
}

func drivesToComponentSlice(ds []*common.Drive) []*rivets.Component {
	comps := []*rivets.Component{}
	for idx, d := range ds {
		serial := d.Serial
		if strings.TrimSpace(serial) == "" {
			serial = strconv.Itoa(idx)
		}
		component := newComponent(common.SlugDrive, d.Vendor, d.Model, serial, d.ProductName,
			&rivets.ComponentAttributes{
				Description:         d.Description,
				ProductName:         d.ProductName,
				Oem:                 d.Oem,
				Metadata:            d.Metadata,
				BusInfo:             d.BusInfo,
				OemID:               d.OemID,
				StorageController:   d.StorageController,
				Protocol:            d.Protocol,
				SmartErrors:         d.SmartErrors,
				SmartStatus:         d.SmartStatus,
				DriveType:           d.Type,
				WWN:                 d.WWN,
				CapacityBytes:       d.CapacityBytes,
				BlockSizeBytes:      d.BlockSizeBytes,
				CapableSpeedGbps:    d.CapableSpeedGbps,
				NegotiatedSpeedGbps: d.NegotiatedSpeedGbps,
				Capabilities:        d.Capabilities,
			},
			d.Status,
			d.Firmware,
		)
		// some drives show up with model numbers in the description field.
		if component.Model == "" && d.Description != "" {
			component.Model = d.Description
		}
		comps = append(comps, component)
	}
	return comps
}

func cpusToComponentSlice(cs []*common.CPU) []*rivets.Component {
	comps := []*rivets.Component{}
	for idx, c := range cs {
		serial := c.Serial
		if strings.TrimSpace(serial) == "" {
			serial = strconv.Itoa(idx)
		}
		component := newComponent(common.SlugCPU, c.Vendor, c.Model, serial, c.ProductName,
			&rivets.ComponentAttributes{
				ID:           c.ID,
				Description:  c.Description,
				ProductName:  c.ProductName,
				Metadata:     c.Metadata,
				Slot:         c.Slot,
				Architecture: c.Architecture,
				ClockSpeedHz: c.ClockSpeedHz,
				Cores:        c.Cores,
				Threads:      c.Threads,
				Capabilities: c.Capabilities,
			},
			c.Status,
			c.Firmware,
		)
		comps = append(comps, component)
	}
	return comps
}

func psusToComponentSlice(ps []*common.PSU) []*rivets.Component {
	comps := []*rivets.Component{}
	for idx, p := range ps {
		serial := p.Serial
		if strings.TrimSpace(serial) == "" {
			serial = strconv.Itoa(idx)
		}
		component := newComponent(common.SlugPSU, p.Vendor, p.Model, serial, p.ProductName,
			&rivets.ComponentAttributes{
				ID:                 p.ID,
				Description:        p.Description,
				ProductName:        p.ProductName,
				PowerCapacityWatts: p.PowerCapacityWatts,
				Oem:                p.Oem,
				Metadata:           p.Metadata,
				Capabilities:       p.Capabilities,
			},
			p.Status,
			p.Firmware,
		)
		comps = append(comps, component)
	}
	return comps
}

func gpusToComponentSlice(gs []*common.GPU) []*rivets.Component {
	comps := []*rivets.Component{}
	for idx, g := range gs {
		serial := g.Serial
		if strings.TrimSpace(serial) == "" {
			serial = strconv.Itoa(idx)
		}
		component := newComponent(common.SlugGPU, g.Vendor, g.Model, serial, g.ProductName,
			&rivets.ComponentAttributes{
				Description:  g.Description,
				ProductName:  g.ProductName,
				Metadata:     g.Metadata,
				Capabilities: g.Capabilities,
			},
			g.Status,
			g.Firmware,
		)
		comps = append(comps, component)
	}
	return comps
}

func cpldsToComponentSlice(cs []*common.CPLD) []*rivets.Component {
	comps := []*rivets.Component{}
	for idx, c := range cs {
		serial := c.Serial
		if strings.TrimSpace(serial) == "" {
			serial = strconv.Itoa(idx)
		}
		component := newComponent(common.SlugCPLD, c.Vendor, c.Model, serial, c.ProductName,
			&rivets.ComponentAttributes{
				Description:  c.Description,
				ProductName:  c.ProductName,
				Metadata:     c.Metadata,
				Capabilities: c.Capabilities,
			},
			c.Status,
			c.Firmware,
		)
		comps = append(comps, component)
	}
	return comps
}

func tpmsToComponentSlice(ts []*common.TPM) []*rivets.Component {
	comps := []*rivets.Component{}
	for idx, t := range ts {
		serial := t.Serial
		if strings.TrimSpace(serial) == "" {
			serial = strconv.Itoa(idx)
		}
		component := newComponent(common.SlugTPM, t.Vendor, t.Model, serial, t.ProductName,
			&rivets.ComponentAttributes{
				Description:   t.Description,
				ProductName:   t.ProductName,
				Metadata:      t.Metadata,
				Capabilities:  t.Capabilities,
				InterfaceType: t.InterfaceType,
			},
			t.Status,
			t.Firmware,
		)
		comps = append(comps, component)
	}
	return comps
}

func storageControllersToComponentSlice(ss []*common.StorageController) []*rivets.Component {
	comps := []*rivets.Component{}
	for idx, s := range ss {
		serial := s.Serial
		if strings.TrimSpace(serial) == "" {
			serial = strconv.Itoa(idx)
		}
		component := newComponent(common.SlugStorageController, s.Vendor, s.Model, serial, s.ProductName,
			&rivets.ComponentAttributes{
				ID:                           s.ID,
				Description:                  s.Description,
				ProductName:                  s.ProductName,
				Oem:                          s.Oem,
				SupportedControllerProtocols: s.SupportedControllerProtocols,
				SupportedDeviceProtocols:     s.SupportedDeviceProtocols,
				SupportedRAIDTypes:           s.SupportedRAIDTypes,
				PhysicalID:                   s.PhysicalID,
				BusInfo:                      s.BusInfo,
				SpeedGbps:                    s.SpeedGbps,
				Metadata:                     s.Metadata,
				Capabilities:                 s.Capabilities,
			},
			s.Status,
			s.Firmware,
		)
		// some controller show up with model numbers in the description field.
		if component.Model == "" && s.Description != "" {
			component.Model = s.Description
		}
		comps = append(comps, component)
	}
	return comps
}

func enclosuresToComponentSlice(es []*common.Enclosure) []*rivets.Component {
	comps := []*rivets.Component{}
	for idx, e := range es {
		serial := e.Serial
		if strings.TrimSpace(serial) == "" {
			serial = strconv.Itoa(idx)
		}
		component := newComponent(common.SlugEnclosure, e.Vendor, e.Model, serial, e.ProductName,
			&rivets.ComponentAttributes{
				ID:           e.ID,
				Description:  e.Description,
				ProductName:  e.ProductName,
				Oem:          e.Oem,
				Metadata:     e.Metadata,
				ChassisType:  e.ChassisType,
				Capabilities: e.Capabilities,
			},
			e.Status,
			e.Firmware,
		)
		comps = append(comps, component)
	}
	return comps
}

func getComponentSlice(device *common.Device) []*rivets.Component {
	components := []*rivets.Component{}

	if device.BIOS != nil {
		components = append(components, biosToComponent(device.BIOS))
	}

	if devBMC := device.BMC; devBMC != nil {
		components = append(components, bmcToComponent(device.BMC))
	}

	if device.Mainboard != nil {
		components = append(components, mainboardToComponent(device.Mainboard))
	}

	if len(device.Memory) > 0 {
		components = append(components, dimmsToComponentSlice(device.Memory)...)
	}

	if len(device.NICs) > 0 {
		components = append(components, nicsToComponentSlice(device.NICs)...)
	}

	if len(device.Drives) > 0 {
		components = append(components, drivesToComponentSlice(device.Drives)...)
	}

	if len(device.PSUs) > 0 {
		components = append(components, psusToComponentSlice(device.PSUs)...)
	}

	if len(device.CPUs) > 0 {
		components = append(components, cpusToComponentSlice(device.CPUs)...)
	}

	if len(device.TPMs) > 0 {
		components = append(components, tpmsToComponentSlice(device.TPMs)...)
	}

	if len(device.CPLDs) > 0 {
		components = append(components, cpldsToComponentSlice(device.CPLDs)...)
	}

	if len(device.GPUs) > 0 {
		components = append(components, gpusToComponentSlice(device.GPUs)...)
	}

	if len(device.StorageControllers) > 0 {
		components = append(components, storageControllersToComponentSlice(device.StorageControllers)...)
	}

	if len(device.Enclosures) > 0 {
		components = append(components, enclosuresToComponentSlice(device.Enclosures)...)
	}

	return components
}
