package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/metal-toolbox/alloy/types"
	"github.com/metal-toolbox/component-inventory/internal/app"
	iconv "github.com/metal-toolbox/component-inventory/internal/inventoryconverter"
	"github.com/metal-toolbox/component-inventory/internal/metrics"
	"github.com/metal-toolbox/component-inventory/internal/version"
	"github.com/metal-toolbox/component-inventory/pkg/api/constants"
	"go.hollow.sh/toolbox/ginauth"
	"go.hollow.sh/toolbox/ginjwt"
	"go.uber.org/zap"
)

var (
	readTimeout  = 10 * time.Second
	writeTimeout = 20 * time.Second

	authMiddleWare *ginauth.MultiTokenMiddleware
	ginNoOp        = func(_ *gin.Context) {}
)

// apiHandler is a function that performs real work for this API.
type apiHandler func(map[string]any) (map[string]any, error)

func composeAppLogging(l *zap.Logger, skippedPaths ...string) gin.HandlerFunc {
	skipMap := map[string]struct{}{}
	for _, path := range skippedPaths {
		skipMap[path] = struct{}{}
	}
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next() // call the next function in the chain
		code := c.Writer.Status()
		metrics.APICallEpilog(start, path, code)

		// only log if we're not skipping this path
		if _, ok := skipMap[path]; ok {
			return
		}

		fields := []zap.Field{
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status-code", code),
			zap.Time("start", start),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.Strings("errors", c.Errors.Errors()))
			l.Error("errors on API request",
				fields...,
			)
			return
		}

		l.Info("api call complete", fields...)
	}
}

// ComposeHTTPServer returns an http.Server that handles our API
func ComposeHTTPServer(theApp *app.App) *http.Server {
	if len(theApp.Cfg.JWTAuth) != 0 {
		var err error
		authMiddleWare, err = ginjwt.NewMultiTokenMiddlewareFromConfigs(theApp.Cfg.JWTAuth...)
		if err != nil {
			theApp.Log.Fatal(
				"failed to initialize auth middleware",
				zap.Error(err),
			)
		}
	}

	g := gin.New()

	if !theApp.Cfg.DeveloperMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// set up common middleware for logging and metrics
	g.Use(composeAppLogging(theApp.Log, constants.LivenessEndpoint), gin.Recovery())

	// some boilerplate setup
	g.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound,
			gin.H{
				"message": "invalid request - route not found",
			},
		)
	})

	// a liveness endpoint
	g.GET(constants.LivenessEndpoint, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"time": time.Now()})
	})

	g.GET(constants.VersionEndpoint, func(c *gin.Context) {
		c.JSON(http.StatusOK, version.Current())
	})

	g.POST("/api/echo",
		composeAuthHandler(createScopes("response")), // auth handler
		wrapAPICall(apiEcho))                         // api function, wrapped into middleware

	g.POST("/api/error",
		composeAuthHandler(createScopes("response")),
		wrapAPICall(apiError))

	// add other API endpoints to the gin Engine as required

	// get the components associated with a server
	g.GET(constants.ComponentsEndpoint+"/:server",
		composeAuthHandler(readScopes("server:component")),
		func(ctx *gin.Context) {
			serverID, err := uuid.Parse(ctx.Param("server"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, map[string]any{
					"message": "invalid server id",
					"err":     err.Error(),
				})
				return
			}

			getInband := true
			qVal, set := ctx.GetQuery("mode")
			if set && qVal == constants.OutOfBandMode {
				getInband = false
			}

			existing, _, err := theApp.FleetDB.GetServerInventory(ctx, serverID, getInband)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, map[string]any{
					"message": "components unavailable",
					"error":   err.Error(),
				})
				return
			}
			// XXX: parse out the components and present them nicely
			ctx.JSON(http.StatusOK, existing.Components)
		})

	// add an API to ingest inventory data
	g.POST(constants.InventoryEndpoint+"/:server",
		composeAuthHandler(updateScopes("server:component")),
		composeInventoryHandler(theApp),
	)

	return &http.Server{
		Addr:         theApp.Cfg.ListenAddress,
		Handler:      g,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
}

// wrapAPICall is an adapter for any arbitrary code so that you can isolate your
// logic from having to take gin-specific data structures and whatnot. It assumes
// your API function takes a map[string]any and returns a JSON-serializable result
// and an error. This function could be altered to pull any kind of parameter out
// of the raw JSON input.
func wrapAPICall(fn apiHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var responseCode int

		m := make(map[string]any)
		if err := ctx.BindJSON(&m); err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"error": err.Error(),
			})
			return
		}

		obj, err := fn(m)
		if err == nil {
			responseCode = http.StatusOK
		} else {
			responseCode = http.StatusInternalServerError
			obj = map[string]any{
				"error": err.Error(),
			}
		}
		ctx.JSON(responseCode, obj)
	}
}

func reject(ctx *gin.Context, code int, msg, err string) {
	ctx.JSON(code, map[string]any{
		"message": msg,
		"err":     err,
	})
}

func composeInventoryHandler(theApp *app.App) gin.HandlerFunc {
	logger := theApp.Log
	fdb := theApp.FleetDB
	return func(ctx *gin.Context) {
		serverID, err := uuid.Parse(ctx.Param("server"))
		if err != nil {
			logger.With(
				zap.Error(err),
				zap.String("server_param", ctx.Param("server")),
			).Warn("bad server id")
			reject(ctx, http.StatusBadRequest, "invalid server id", err.Error())
			return
		}

		inband := true
		qVal, set := ctx.GetQuery("mode")
		if set && qVal == constants.OutOfBandMode {
			inband = false
		}

		logger.With(
			zap.String("server.id", serverID.String()),
			zap.Bool("inband", inband),
		).Debug("processing inventory")

		var dev types.InventoryDevice
		if err = ctx.BindJSON(&dev); err != nil {
			logger.With(
				zap.Error(err),
			).Warn("bad server payload")
			reject(ctx, http.StatusBadRequest, "invalid server inventory", err.Error())
			return
		}

		if dev.Inv == nil {
			logger.Warn("empty inventory")
			reject(ctx, http.StatusBadRequest, "empty inventory", "")
			return
		}

		existing, _, err := fdb.GetServerInventory(ctx, serverID, inband)
		if err != nil {
			logger.With(zap.Error(err)).Warn("server lookup")
			reject(ctx, http.StatusBadRequest, "unable to retrieve server", err.Error())
			return
		}

		latest := iconv.ToRivetsServer(existing.Name, existing.Facility, dev.Inv, dev.BiosCfg)
		// sanity check the latest to what exists in FleetDB
		compareComponents(existing, latest, logger)

		_, err = fdb.SetServerInventory(ctx, serverID, latest, inband)
		if err != nil {
			logger.With(
				zap.Error(err),
			).Warn("server update to fleet db")
			reject(ctx, http.StatusInternalServerError, "unable to process inventory", err.Error())
			return
		}

		ctx.Status(http.StatusCreated)
	}
}

func composeAuthHandler(scopes []string) gin.HandlerFunc {
	if authMiddleWare == nil {
		return ginNoOp
	}
	return authMiddleWare.AuthRequired(scopes)
}

func createScopes(items ...string) []string {
	s := []string{"write", "create"}
	for _, i := range items {
		s = append(s, fmt.Sprintf("create:%s", i))
	}

	return s
}

//nolint:unused
func readScopes(items ...string) []string {
	s := []string{"read"}
	for _, i := range items {
		s = append(s, fmt.Sprintf("read:%s", i))
	}

	return s
}

//nolint:unused
func updateScopes(items ...string) []string {
	s := []string{"write", "update"}
	for _, i := range items {
		s = append(s, fmt.Sprintf("update:%s", i))
	}

	return s
}

//nolint:unused
func deleteScopes(items ...string) []string {
	s := []string{"write", "delete"}
	for _, i := range items {
		s = append(s, fmt.Sprintf("delete:%s", i))
	}

	return s
}
