package constants

const (
	LivenessEndpoint           = "/_health/liveness"
	VersionEndpoint            = "/api/version"
	ComponentsEndpoint         = "/components"
	InbandInventoryEndpoint    = "/inventory/in-band"
	OutofbandInventoryEndpoint = "/inventory/out-of-band"

	// server service attribute to look up the BMC IP Address in
	BmcAttributeNamespace = "sh.hollow.bmc_info"

	// server server service BMC address attribute key found under the bmcAttributeNamespace
	BmcIPAddressAttributeKey = "address"

	// serverservice namespace prefix the data is stored in.
	ServerServiceNSPrefix = "sh.hollow.alloy"

	// server vendor, model attributes are stored in this namespace.
	ServerVendorAttributeNS = ServerServiceNSPrefix + ".server_vendor_attributes"

	// additional server metadata are stored in this namespace.
	ServerMetadataAttributeNS = ServerServiceNSPrefix + ".server_metadata_attributes"

	// errors that occurred when connecting/collecting inventory from the bmc are stored here.
	ServerBMCErrorsAttributeNS = ServerServiceNSPrefix + ".server_bmc_errors"

	// server service server serial attribute key
	ServerSerialAttributeKey = "serial"

	// server service server model attribute key
	ServerModelAttributeKey = "model"

	// server service server vendor attribute key
	ServerVendorAttributeKey = "vendor"
)
