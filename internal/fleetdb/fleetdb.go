package internalfleetdb

import (
	"context"
	"net/url"
	"time"

	common "github.com/bmc-toolbox/common"
	oidc "github.com/coreos/go-oidc"
	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/metal-toolbox/component-inventory/internal/app"
	"github.com/metal-toolbox/component-inventory/internal/inventoryconverter"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	rivets "github.com/metal-toolbox/rivets/types"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type Client interface {
	GetComponents(context.Context, uuid.UUID, *fleetdb.PaginationParams) (fleetdb.ServerComponentSlice, *fleetdb.ServerResponse, error)
	GetServerInventory(context.Context, uuid.UUID, bool) (*rivets.Server, *fleetdb.ServerResponse, error)
	UpdateServerInventory(context.Context, uuid.UUID, *rivets.Server, bool, *zap.Logger) error
	GetInventoryConverter() *inventoryconverter.InventoryConverter
}

// connectionTimeout is the maximum amount of time spent on each http connection to FleetDBClient.
var connectionTimeout = 30 * time.Second

// Creates a new Client, with reasonable defaults
func NewFleetDBClient(ctx context.Context, cfg *app.Configuration) (Client, error) {
	fleetDBOpts := &cfg.FleetDBAPIOptions
	client, err := getFleetDBAPIClient(ctx, fleetDBOpts)

	if err != nil {
		return nil, err
	}

	slugs := make(map[string]bool)
	serverComponentTypes := common.ComponentTypes()
	for _, ct := range serverComponentTypes {
		slugs[ct] = true
	}

	return &fleetDBClient{
		client:                     client,
		inventoryConverterInstance: inventoryconverter.NewInventoryConverter(slugs),
	}, nil
}

func getFleetDBAPIClient(ctx context.Context, cfg *app.FleetDBAPIOptions) (*fleetdb.Client, error) {
	if cfg.DisableOAuth {
		return fleetdb.NewClientWithToken("fake", cfg.Endpoint, nil)
	}

	// init retryable http client
	retryableClient := retryablehttp.NewClient()

	// set retryable HTTP client to be the otel http client to collect telemetry
	retryableClient.HTTPClient = otelhttp.DefaultClient

	// setup oidc provider
	provider, err := oidc.NewProvider(ctx, cfg.OidcIssuerEndpoint)
	if err != nil {
		return nil, err
	}

	clientID := "component-inventory"

	if cfg.OidcClientID != "" {
		clientID = cfg.OidcClientID
	}

	// setup oauth configuration
	oauthConfig := clientcredentials.Config{
		ClientID:       clientID,
		ClientSecret:   cfg.OidcClientSecret,
		TokenURL:       provider.Endpoint().TokenURL,
		Scopes:         cfg.OidcClientScopes,
		EndpointParams: url.Values{"audience": []string{cfg.OidcAudienceEndpoint}},
		// with this the oauth client spends less time identifying the client grant mechanism.
		AuthStyle: oauth2.AuthStyleInParams,
	}

	// wrap OAuth transport, cookie jar in the retryable client
	oAuthclient := oauthConfig.Client(ctx)

	retryableClient.HTTPClient.Transport = oAuthclient.Transport
	retryableClient.HTTPClient.Jar = oAuthclient.Jar

	httpClient := retryableClient.StandardClient()
	httpClient.Timeout = connectionTimeout

	return fleetdb.NewClientWithToken(
		cfg.OidcClientSecret,
		cfg.Endpoint,
		httpClient,
	)
}

type fleetDBClient struct {
	client                     *fleetdb.Client
	inventoryConverterInstance *inventoryconverter.InventoryConverter
}

func (fc fleetDBClient) GetServerInventory(ctx context.Context, id uuid.UUID, inband bool) (*rivets.Server, *fleetdb.ServerResponse, error) {
	return fc.client.GetServerInventory(ctx, id, inband)
}

func (fc fleetDBClient) GetComponents(ctx context.Context, id uuid.UUID, params *fleetdb.PaginationParams) (fleetdb.ServerComponentSlice, *fleetdb.ServerResponse, error) {
	return fc.client.GetComponents(ctx, id, params)
}

func (fc fleetDBClient) UpdateServerInventory(ctx context.Context, serverID uuid.UUID, rivetsServer *rivets.Server, inband bool, log *zap.Logger) error {
	if _, err := fc.client.SetServerInventory(ctx, serverID, rivetsServer, inband); err != nil {
		log.Error("set inventory fail", zap.String("server", serverID.String()), zap.String("err", err.Error()))
		return err
	}
	return nil
}

func (fc fleetDBClient) GetInventoryConverter() *inventoryconverter.InventoryConverter {
	return fc.inventoryConverterInstance
}
