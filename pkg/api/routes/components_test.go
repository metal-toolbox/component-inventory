package routes

import (
	"github.com/metal-toolbox/component-inventory/internal/app"

	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var serverUUID = uuid.New()

var validComponents = fleetdb.ServerComponentSlice{
	{
		ServerUUID:        serverUUID,
		Vendor:            "vendor1",
		Model:             "model1",
		Serial:            "xyz123",
		ComponentTypeSlug: "slug1",
	},
	{
		ServerUUID:        serverUUID,
		Vendor:            "vendor2",
		Model:             "slug2-model1",
		Serial:            "foobar2",
		ComponentTypeSlug: "slug2",
	},
	{
		ServerUUID:        serverUUID,
		Vendor:            "vendor3",
		Model:             "slug2-model2",
		Serial:            "fizzbuzz",
		ComponentTypeSlug: "slug2",
	},
}

func getComponentsHandler(t *testing.T, comps *fleetdb.ServerComponentSlice, code int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var byt []byte
		if comps != nil {
			var err error
			srvResponse := fleetdb.ServerResponse{
				Records: comps,
			}
			byt, err = json.Marshal(srvResponse)
			if err != nil {
				t.Fatalf("serializing server response: %s", err.Error())
			}
		}

		w.WriteHeader(code)
		if byt != nil {
			_, err := w.Write(byt)
			if err != nil {
				t.Fatalf("writing http response: %s", err.Error())
			}
		}
	}
}

func TestFetchServerComponents(t *testing.T) {
	t.Parallel()
	t.Run("valid component return", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		// XXX: go 1.22 changes how you pattern match in ServeMux
		mux.HandleFunc(
			// this is for test-debugging. trying to figure out that the client is mis-handling the URL
			// hurts otherwise.
			"/",
			func(w http.ResponseWriter, r *http.Request) {
				t.Logf("route not found: %s", r.URL.Path)
				w.WriteHeader(http.StatusBadRequest)
			},
		)
		mux.HandleFunc(
			fmt.Sprintf("/api/v1/servers/%s/components", serverUUID),
			getComponentsHandler(t, &validComponents, 200),
		)

		ts := httptest.NewServer(mux)

		logger := app.GetLogger(true)

		client := getFleetDBClient(&app.Configuration{
			FleetDBAddress: ts.URL,
		})

		result, err := fetchServerComponents(client, serverUUID, logger)
		require.NoError(t, err)
		require.Equal(t, 2, len(result))
		require.Equal(t, 2, len(result["slug2"]))
	})
	t.Run("server error", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		mux.HandleFunc(
			"/",
			func(w http.ResponseWriter, r *http.Request) {
				t.Logf("route not found: %s", r.URL.Path)
				w.WriteHeader(http.StatusBadRequest)
			},
		)
		mux.HandleFunc(
			fmt.Sprintf("/api/v1/servers/%s/components", serverUUID),
			getComponentsHandler(t, nil, 500),
		)

		ts := httptest.NewServer(mux)

		logger := app.GetLogger(true)

		client := getFleetDBClient(&app.Configuration{
			FleetDBAddress: ts.URL,
		})

		_, err := fetchServerComponents(client, serverUUID, logger)
		require.Error(t, err)
		var srvErr fleetdb.ServerError
		require.ErrorAs(t, err, &srvErr, "unexpected error type")
		require.Equal(t, 500, srvErr.StatusCode)
	})
}
