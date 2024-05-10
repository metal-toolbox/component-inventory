package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/hashicorp/go-retryablehttp"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	rootCmd "github.com/metal-toolbox/component-inventory/cmd"
	"github.com/metal-toolbox/component-inventory/internal/app"
	"github.com/metal-toolbox/component-inventory/internal/metrics"
	"github.com/metal-toolbox/component-inventory/internal/version"
	"github.com/metal-toolbox/component-inventory/pkg/api/routes"
	"github.com/spf13/cobra"
)

const (
	dialTimeout     = 30 * time.Second
	shutdownTimeout = 10 * time.Second
)

func getFleetDBClient(cfg *app.Configuration) (*fleetdb.Client, error) {
	if cfg.FleetDBOpts.DisableOAuth {
		return fleetdb.NewClient(cfg.FleetDBOpts.Endpoint, nil)
	}

	ctx := context.Background()

	// init retryable http client
	retryableClient := retryablehttp.NewClient()

	// set retryable HTTP client to be the otel http client to collect telemetry
	retryableClient.HTTPClient = otelhttp.DefaultClient

	// setup oidc provider
	provider, err := oidc.NewProvider(ctx, cfg.FleetDBOpts.IssuerEndpoint)
	if err != nil {
		return nil, err
	}

	clientID := "component-inventory"

	if cfg.FleetDBOpts.ClientID != "" {
		clientID = cfg.FleetDBOpts.ClientID
	}

	// setup oauth configuration
	oauthConfig := clientcredentials.Config{
		ClientID:       clientID,
		ClientSecret:   cfg.FleetDBOpts.ClientSecret,
		TokenURL:       provider.Endpoint().TokenURL,
		Scopes:         cfg.FleetDBOpts.ClientScopes,
		EndpointParams: url.Values{"audience": []string{cfg.FleetDBOpts.AudienceEndpoint}},
		// with this the oauth client spends less time identifying the client grant mechanism.
		AuthStyle: oauth2.AuthStyleInParams,
	}

	// wrap OAuth transport, cookie jar in the retryable client
	oAuthclient := oauthConfig.Client(ctx)

	retryableClient.HTTPClient.Transport = oAuthclient.Transport
	retryableClient.HTTPClient.Jar = oAuthclient.Jar

	httpClient := retryableClient.StandardClient()
	httpClient.Timeout = dialTimeout

	return fleetdb.NewClientWithToken(
		cfg.FleetDBOpts.ClientSecret,
		cfg.FleetDBOpts.Endpoint,
		httpClient,
	)
}

// install server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run API service",
	Run: func(c *cobra.Command, args []string) {
		cfg, err := app.LoadConfiguration(rootCmd.CfgFile)
		if err != nil {
			log.Fatalf("loading configuration: %s", err.Error())
		}

		logger := app.GetLogger(cfg.DeveloperMode)
		//nolint:errcheck
		defer logger.Sync()

		fdb, err := getFleetDBClient(cfg)
		if err != nil {
			logger.With(
				zap.Error(err),
			).Fatal("creating fleetdb client")
		}

		ctx, appCancel := context.WithCancel(c.Context())
		app := app.NewApp(ctx, cfg, logger, fdb)

		metrics.ListenAndServe()

		// the ignored parameter here is a context annotated with otel-init-go configuration
		_, otelShutdown := otelinit.InitOpenTelemetry(c.Context(), "cis-api-server")

		logger.Info("app initialized",
			zap.String("version", version.Current().String()),
		)

		srv := routes.ComposeHTTPServer(app)
		go func() {
			if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("error serving API",
					zap.Error(err),
				)
			}
		}()

		app.WaitForSignal()
		logger.Info("signaled to terminate")
		appCancel()

		// call server shutdown with timeout
		ctx, cancel := context.WithTimeout(c.Context(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatal("server shutdown error",
				zap.Error(err),
			)
		}
		otelShutdown(ctx)
		logger.Info("OK, done.")
	},
}

// install command flags
func init() {
	rootCmd.RootCmd.AddCommand(serverCmd)
}
