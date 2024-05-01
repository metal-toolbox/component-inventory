package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/metal-toolbox/component-inventory/internal/app"
	"github.com/stretchr/testify/require"

	common "github.com/bmc-toolbox/common"
	internalfleetdb "github.com/metal-toolbox/component-inventory/internal/fleetdb"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	rivets "github.com/metal-toolbox/rivets/types"
	"go.uber.org/zap"
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

var validComponentTypes = fleetdb.ServerComponentTypeSlice{
	&fleetdb.ServerComponentType{
		ID:   "02dc2503-b64c-439b-9f25-8e130705f14a",
		Name: "Backplane-Expander",
		Slug: "backplane-expander",
	},
	&fleetdb.ServerComponentType{
		ID:   "1e0c3417-d63c-4fd5-88f7-4c525c70da12",
		Name: "Mainboard",
		Slug: "mainboard",
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

func getComponentTypesHandler(t *testing.T, comps *fleetdb.ServerComponentSlice, code int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var byt []byte
		if comps != nil {
			var err error
			srvResponse := fleetdb.ServerResponse{
				Records: validComponentTypes,
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
		mux.HandleFunc("/api/v1/server-component-types",
			getComponentTypesHandler(t, &validComponents, 200),
		)
		mux.HandleFunc(
			fmt.Sprintf("/api/v1/servers/%s/components", serverUUID),
			getComponentsHandler(t, &validComponents, 200),
		)

		ts := httptest.NewServer(mux)

		logger := app.GetLogger(true)

		client, err := internalfleetdb.NewFleetDBClient(context.Background(), &app.Configuration{
			FleetDBAddress: ts.URL,
		})
		require.NoError(t, err)

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
		mux.HandleFunc("/api/v1/server-component-types",
			getComponentTypesHandler(t, &validComponents, 200),
		)
		mux.HandleFunc(
			fmt.Sprintf("/api/v1/servers/%s/components", serverUUID),
			getComponentsHandler(t, nil, 500),
		)

		ts := httptest.NewServer(mux)

		logger := app.GetLogger(true)

		client, err := internalfleetdb.NewFleetDBClient(context.Background(), &app.Configuration{
			FleetDBAddress: ts.URL,
		})
		require.NoError(t, err)

		_, err = fetchServerComponents(client, serverUUID, logger)
		require.Error(t, err)
		var srvErr fleetdb.ServerError
		require.ErrorAs(t, err, &srvErr, "unexpected error type")
		require.Equal(t, 500, srvErr.StatusCode)
	})
}

func TestCompareComponents(t *testing.T) {
	testCases := []struct {
		desc        string
		server      *rivets.Server
		dev         *rivets.Server
		expectMatch bool
	}{
		{
			"2 servers nil component match case",
			&rivets.Server{
				Name:       "slug1",
				Components: nil,
				ID:         "123",
			},
			&rivets.Server{
				Name:       "slug2",
				Components: nil,
				ID:         "124",
			},
			true,
		},
		{
			"2 servers components match case",
			&rivets.Server{
				Name: "server1",
				Components: []*rivets.Component{
					{
						Name: "BIOS",
						Firmware: &common.Firmware{
							Installed: "1.2.3",
						},
						Vendor: "Dell",
						Model:  "model1",
					},
				},
				ID: "123",
			},
			&rivets.Server{
				Name: "server2",
				Components: []*rivets.Component{
					{
						Name: "BIOS",
						Firmware: &common.Firmware{
							Installed: "1.2.3",
						},
						Vendor: "Dell",
						Model:  "model1",
					},
				},
				ID: "124",
			},
			true,
		},
		{
			"2 servers components not match case",
			&rivets.Server{
				Name: "server1",
				Components: []*rivets.Component{
					{
						Name: "BIOS",
						Firmware: &common.Firmware{
							Installed: "1.2.3",
						},
						Vendor: "Dell",
						Model:  "model1",
					},
				},
				ID: "123",
			},
			&rivets.Server{
				Name: "server2",
				Components: []*rivets.Component{
					{
						Name: "BIOS",
						Firmware: &common.Firmware{
							Installed: "1.2.4",
						},
						Vendor: "Dell",
						Model:  "model",
					},
				},
				ID: "124",
			},
			false,
		},
		{
			"expected alloy server nil components not match case",
			&rivets.Server{
				Name: "server1",
				Components: []*rivets.Component{
					{
						Name: "BIOS",
						Firmware: &common.Firmware{
							Installed: "1,2,3",
						},
						Vendor: "Dell",
						Model:  "model1",
					},
				},
				ID: "123",
			},
			&rivets.Server{
				Name:       "server2",
				Components: nil,
				ID:         "124",
			},
			false,
		},
	}

	for _, tc := range testCases {
		if got := compareComponents(tc.server, tc.dev, zap.NewNop()); got != tc.expectMatch {
			t.Errorf("test %v got compare result %v, expect %v", tc.desc, got, tc.expectMatch)
		}
	}

}
