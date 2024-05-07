package app

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// XXX: be careful here. Compound names need to be valid prometheus metric names (used in internal/metrics.go)
const AppName = "component_inventory"

type App struct {
	Log  *zap.Logger
	Cfg  *Configuration
	ctx  context.Context
	term <-chan os.Signal
	opts map[string]any
}

// Option provides a path for adding arbitrary stuff to an App.
type Option func(*App)

// New Option composes a generic Option for an App.
func NewOption(key string, opt any) Option {
	return func(a *App) {
		a.opts[key] = opt
	}
}

// NewApp composes the provided Configuration and Logger into a new App object
func NewApp(ctx context.Context, cfg *Configuration, log *zap.Logger, opts ...Option) *App {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	app := &App{
		Log:  log,
		Cfg:  cfg,
		ctx:  ctx,
		term: termChan,
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

// WaitForSignal blocks on the Server's internal signal channel until we catch SIGTERM or SIGINT
func (a *App) WaitForSignal() {
	<-a.term
}

// ContextDone indicates whether an App's internal context has expired or been canceled
// We cancel the internal context on SIGTERM or SIGINT to signal anything interested that
// it's time to go.
func (a *App) ContextDone() bool {
	return a.ctx.Err() != nil
}

// LogRunningConfig does exactly what it says on the tin. It is only a side-effect.
func (a *App) LogRunningConfig() {
	a.Log.Info("running configuration",
		zap.String("fleetdb.address", a.Cfg.FleetDBAPIOptions.Endpoint),
		zap.String("listen.address", a.Cfg.ListenAddress),
		zap.Bool("developer.mode", a.Cfg.DeveloperMode),
		// do something for the JWTAuthConfig
	)
}

// LoadConfiguration opens and parses the configuration file and then applies any
// environmental overrides
func LoadConfiguration(cfgFile string) (*Configuration, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix(AppName)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	cfg := &Configuration{}

	fh, err := os.Open(cfgFile)
	if err != nil {
		return nil, errors.Wrap(err, "opening config file "+cfgFile)
	}

	if err = v.ReadConfig(fh); err != nil {
		return nil, errors.Wrap(err, "reading config "+cfgFile)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "unmarshaling config")
	}

	// for injected overrides like secrets
	if err := envVarOverrides(v, cfg); err != nil {
		return nil, errors.Wrap(err, "configuring environment orverrides")
	}

	if cfg.ListenAddress == "" {
		return nil, errors.New("listen address not set")
	}

	if cfg.FleetDBAPIOptions.Endpoint == "" {
		return nil, errors.New("fleetdb endpoint not set")
	}

	return cfg, nil
}

// nolint:gocyclo // parameter validation is cyclomatic
func envVarOverrides(v *viper.Viper, cfg *Configuration) error {
	if addr := v.GetString("listen.address"); addr != "" {
		cfg.ListenAddress = addr
	}

	if v.GetBool("developer.mode") {
		cfg.DeveloperMode = true
	}

	// sanity checks
	if cfg.ListenAddress == "" {
		return errors.New("no listen address set")
	}

	if v.GetString("fleetdb.disable.oauth") != "" {
		cfg.FleetDBAPIOptions.DisableOAuth = v.GetBool("fleetdb.disable.oauth")
	}

	if cfg.FleetDBAPIOptions.DisableOAuth {
		return nil
	}

	if v.GetString("fleetdb.oidc.issuer.endpoint") != "" {
		cfg.FleetDBAPIOptions.OidcIssuerEndpoint = v.GetString("fleetdb.oidc.issuer.endpoint")
	}

	if cfg.FleetDBAPIOptions.OidcAudienceEndpoint == "" {
		return errors.New("fleetdb oidc.audience.endpoint not defined")
	}

	if v.GetString("fleetdb.oidc.audience.endpoint") != "" {
		cfg.FleetDBAPIOptions.OidcAudienceEndpoint = v.GetString("fleetdb.oidc.audience.endpoint")
	}

	if cfg.FleetDBAPIOptions.OidcAudienceEndpoint == "" {
		return errors.New("fleetdb oidc.audience.endpoint not defined")
	}

	if v.GetString("fleetdb.oidc.client.secret") != "" {
		cfg.FleetDBAPIOptions.OidcClientSecret = v.GetString("fleetdb.oidc.client.secret")
	}

	if cfg.FleetDBAPIOptions.OidcClientSecret == "" {
		return errors.New("fleetdb oidc.client.secret not defined")
	}

	if v.GetString("fleetdb.oidc.client.id") != "" {
		cfg.FleetDBAPIOptions.OidcClientID = v.GetString("fleetdb.oidc.client.id")
	}

	if cfg.FleetDBAPIOptions.OidcClientID == "" {
		return errors.New("fleetdb oidc.client.id not defined")
	}

	if v.GetString("fleetdb.oidc.client.scopes") != "" {
		cfg.FleetDBAPIOptions.OidcClientScopes = v.GetStringSlice("fleetdb.oidc.client.scopes")
	}

	if len(cfg.FleetDBAPIOptions.OidcClientScopes) == 0 {
		return errors.New("fleetdb oidc.client.scopes not defined")
	}

	return nil
}

// GetLogger constructs a new logger for composition within an App
func GetLogger(dev bool) *zap.Logger {
	if dev {
		return zap.Must(zap.NewDevelopment(
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		))
	}
	return zap.Must(zap.NewProduction(
		zap.AddCaller(),
	))
}
