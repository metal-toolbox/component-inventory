package app

import (
	"go.hollow.sh/toolbox/ginjwt"
)

type Configuration struct {
	ListenAddress     string              `mapstructure:"listen_address"`
	DeveloperMode     bool                `mapstructure:"developer_mode"`
	JWTAuth           []ginjwt.AuthConfig `mapstructure:"ginjwt_auth"`
	FleetDBAPIOptions FleetDBAPIOptions   `mapstructure:"fleetdb"`
}

// https://github.com/metal-toolbox/fleetdb
type FleetDBAPIOptions struct {
	Endpoint             string   `mapstructure:"endpoint"`
	OidcIssuerEndpoint   string   `mapstructure:"oidc_issuer_endpoint"`
	OidcAudienceEndpoint string   `mapstructure:"oidc_audience_endpoint"`
	OidcClientSecret     string   `mapstructure:"oidc_client_secret"`
	OidcClientID         string   `mapstructure:"oidc_client_id"`
	OidcClientScopes     []string `mapstructure:"oidc_client_scopes"`
	DisableOAuth         bool     `mapstructure:"disable_oauth"`
}
