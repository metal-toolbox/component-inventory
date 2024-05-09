package app

import (
	"go.hollow.sh/toolbox/ginjwt"
)

type Configuration struct {
	ListenAddress string              `mapstructure:"listen_address"`
	DeveloperMode bool                `mapstructure:"developer_mode"`
	JWTAuth       []ginjwt.AuthConfig `mapstructure:"ginjwt_auth"`
	FleetDBOpts   FleetDBAPIOptions   `mapstructure:"fleetdb"`
}

// https://github.com/metal-toolbox/fleetdb
type FleetDBAPIOptions struct {
	Endpoint         string   `mapstructure:"endpoint"`
	DisableOAuth     bool     `mapstructure:"disable_oauth"`
	AudienceEndpoint string   `mapstructure:"audience_endpoint"`
	ClientID         string   `mapstructure:"client_id"`
	ClientSecret     string   `mapstructure:"client_secret"`
	ClientScopes     []string `mapstructure:"client_scopes"`
}
