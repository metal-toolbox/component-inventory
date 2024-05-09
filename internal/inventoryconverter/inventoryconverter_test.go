package inventoryconverter

// if a lot of these tests look similar, they are. The biggest computation in the translation
// from bmc-toolbox to rivets data-structures is the generation of a missing serial number. A
// few mutations include assigning model from description and those are tested as well.

import (
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/stretchr/testify/require"
)

func TestBIOSToComponent(t *testing.T) {
	t.Parallel()
	bios := &common.BIOS{
		Common: common.Common{
			Vendor: "ABC Computer",
		},
	}
	got := biosToComponent(bios)
	require.Equal(t, "0", got.Serial)
}

func TestBMCToComponent(t *testing.T) {
	t.Parallel()
	bmc := &common.BMC{
		Common: common.Common{
			Vendor: "fizzbuzz, Inc.",
		},
	}
	got := bmcToComponent(bmc)
	require.Equal(t, "0", got.Serial)
}

func TestMainboardToComponent(t *testing.T) {
	t.Parallel()
	mb := &common.Mainboard{
		Common: common.Common{
			Vendor: "Shenzen Mobos",
		},
	}
	got := mainboardToComponent(mb)
	require.Equal(t, "0", got.Serial)
}

func TestDimmsToComponentSlice(t *testing.T) {
	t.Parallel()
	dimms := []*common.Memory{
		&common.Memory{
			Slot: "DIMM.Socket.1",
		},
		&common.Memory{
			Common: common.Common{
				Serial: "abc123",
			},
			Slot: "4",
		},
	}
	got := dimmsToComponentSlice(dimms)
	require.Equal(t, 2, len(got))
	require.Equal(t, "1", got[0].Attributes.Slot)
	require.Equal(t, "0", got[0].Serial)
	require.Equal(t, "4", got[1].Attributes.Slot)
	require.Equal(t, "abc123", got[1].Serial)
}

func TestNICsToComponentSlice(t *testing.T) {
	t.Parallel()
	nics := []*common.NIC{
		&common.NIC{},
		&common.NIC{},
	}
	got := nicsToComponentSlice(nics)
	require.Equal(t, 2, len(got))
	require.Equal(t, "0", got[0].Serial)
	require.Equal(t, "1", got[1].Serial)
}

func TestDrivesToComponentSlice(t *testing.T) {
	t.Parallel()
	drives := []*common.Drive{
		&common.Drive{
			Common: common.Common{
				Description: "cool-drive",
			},
		},
		&common.Drive{
			Common: common.Common{
				Model: "cool-drive ultra",
			},
		},
	}
	got := drivesToComponentSlice(drives)
	require.Equal(t, 2, len(got))
	require.Equal(t, "0", got[0].Serial)
	require.Equal(t, "cool-drive", got[0].Model)
	require.Equal(t, "1", got[1].Serial)
	require.Equal(t, "cool-drive ultra", got[1].Model)
}

func TestCPUsToComponentSlice(t *testing.T) {
	t.Parallel()
	cpus := []*common.CPU{
		&common.CPU{},
		&common.CPU{},
	}
	got := cpusToComponentSlice(cpus)
	require.Equal(t, 2, len(got))
	require.Equal(t, "0", got[0].Serial)
	require.Equal(t, "1", got[1].Serial)
}

func TestPSUsToComponentSlice(t *testing.T) {
	t.Parallel()
	psus := []*common.PSU{
		&common.PSU{},
		&common.PSU{},
	}
	got := psusToComponentSlice(psus)
	require.Equal(t, 2, len(got))
	require.Equal(t, "0", got[0].Serial)
	require.Equal(t, "1", got[1].Serial)
}

func TestGPUsToComponentSlice(t *testing.T) {
	t.Parallel()
	gpus := []*common.GPU{
		&common.GPU{},
		&common.GPU{},
	}
	got := gpusToComponentSlice(gpus)
	require.Equal(t, 2, len(got))
	require.Equal(t, "0", got[0].Serial)
	require.Equal(t, "1", got[1].Serial)
}

func TestCPLDsToComponentSlice(t *testing.T) {
	t.Parallel()
	cplds := []*common.CPLD{
		&common.CPLD{},
		&common.CPLD{},
	}
	got := cpldsToComponentSlice(cplds)
	require.Equal(t, 2, len(got))
	require.Equal(t, "0", got[0].Serial)
	require.Equal(t, "1", got[1].Serial)
}

func TestTPMsToComponentSlice(t *testing.T) {
	t.Parallel()
	tpms := []*common.TPM{
		&common.TPM{},
		&common.TPM{},
	}
	got := tpmsToComponentSlice(tpms)
	require.Equal(t, 2, len(got))
	require.Equal(t, "0", got[0].Serial)
	require.Equal(t, "1", got[1].Serial)
}

func TestStorageControllersToComponentSlice(t *testing.T) {
	t.Parallel()
	scs := []*common.StorageController{
		&common.StorageController{
			Common: common.Common{
				Description: "Controlmaster",
			},
		},
		&common.StorageController{
			Common: common.Common{
				Model: "Controlmaster II",
			},
		},
	}
	got := storageControllersToComponentSlice(scs)
	require.Equal(t, 2, len(got))
	require.Equal(t, "0", got[0].Serial)
	require.Equal(t, "Controlmaster", got[0].Model)
	require.Equal(t, "1", got[1].Serial)
	require.Equal(t, "Controlmaster II", got[1].Model)
}

func TestEnclosuresToComponentSlice(t *testing.T) {
	t.Parallel()
	encs := []*common.Enclosure{
		&common.Enclosure{},
		&common.Enclosure{},
	}
	got := enclosuresToComponentSlice(encs)
	require.Equal(t, 2, len(got))
	require.Equal(t, "0", got[0].Serial)
	require.Equal(t, "1", got[1].Serial)
}
