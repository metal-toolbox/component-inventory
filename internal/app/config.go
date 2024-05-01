package app

import (
	"go.hollow.sh/toolbox/ginjwt"
)

type Configuration struct {
	ListenAddress  string              `mapstructure:"listen_address"`
	DeveloperMode  bool                `mapstructure:"developer_mode"`
	JWTAuth        []ginjwt.AuthConfig `mapstructure:"ginjwt_auth"`
	FleetDBAddress string              `mapstructure:"fleetdb_address"`
	FleetDBToken   string              `mapstructure:"fleetdb_token"`
}
