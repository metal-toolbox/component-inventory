package constants

const (
	LivenessEndpoint           = "/_health/liveness"
	VersionEndpoint            = "/api/version"
	ComponentsEndpoint         = "/components"
	InbandInventoryEndpoint    = "/inventory/in-band"
	OutofbandInventoryEndpoint = "/inventory/out-of-band"

	// fleetdb attribute to look up the BMC IP Address in
	BmcAttributeNamespace = "sh.hollow.bmc_info"

	// server fleetdb BMC address attribute key found under the bmcAttributeNamespace
	BmcIPAddressAttributeKey = "address"

	// fleetdb namespace prefix the data is stored in.
	FleetDBNSPrefix = "sh.hollow.alloy"

	// server vendor, model attributes are stored in this namespace.
	ServerVendorAttributeNS = FleetDBNSPrefix + ".server_vendor_attributes"

	// additional server metadata are stored in this namespace.
	ServerMetadataAttributeNS = FleetDBNSPrefix + ".server_metadata_attributes"

	// errors that occurred when connecting/collecting inventory from the bmc are stored here.
	ServerBMCErrorsAttributeNS = FleetDBNSPrefix + ".server_bmc_errors"

	// fleetdb server serial attribute key
	ServerSerialAttributeKey = "serial"

	// fleetdb server model attribute key
	ServerModelAttributeKey = "model"

	// fleetdb server vendor attribute key
	ServerVendorAttributeKey = "vendor"
)
