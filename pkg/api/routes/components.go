package routes

import (
	"context"
	"time"

	"github.com/google/uuid"
	internalfleetdb "github.com/metal-toolbox/component-inventory/internal/fleetdb"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	rdb "github.com/metal-toolbox/rivets/fleetdb"
	rivets "github.com/metal-toolbox/rivets/types"
	"go.uber.org/zap"
)

var fleetDBTimeout = 3 * time.Minute

// this is a map of "component_type_name" to the actual inventory data for each component
type serverComponents map[string][]*rivets.Component

func fetchServerComponents(client internalfleetdb.Client, srvid uuid.UUID, l *zap.Logger) (serverComponents, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fleetDBTimeout)
	defer cancel()

	fleetDBcomps, _, err := client.GetComponents(ctx, srvid, &fleetdb.PaginationParams{})
	if err != nil {
		return nil, err
	}

	comps := make(map[string][]*rivets.Component)

	for _, c := range fleetDBcomps { //nolint
		c := c
		cPtr, err := rdb.RecordToComponent(&c)
		if err != nil {
			l.Warn("error converting component",
				zap.String("component_name", c.ComponentTypeSlug),
				zap.String("server_id", c.ServerUUID.String()),
			)
			continue
		}
		cAry, ok := comps[cPtr.Name] // cPtr.Name is the record's ComponentTypeSlug
		if !ok {
			cAry = make([]*rivets.Component, 0, len(fleetDBcomps))
		}
		cAry = append(cAry, cPtr)
		comps[cPtr.Name] = cAry
	}
	return comps, nil
}

// compareComponents compares components between two rivets.Server.
// It logs differences and return false if there are differences.
// The return value is used for testing only.
func compareComponents(fleetServer, alloyServer *rivets.Server, log *zap.Logger) bool {
	match := true
	alloyComponentsMap := make(map[string][]*rivets.Component)
	for _, component := range alloyServer.Components {
		slug := component.Name
		if _, ok := alloyComponentsMap[slug]; !ok {
			alloyComponentsMap[slug] = make([]*rivets.Component, 0)
		}
		alloyComponentsMap[slug] = append(alloyComponentsMap[slug], component)
	}

	for _, fleetComponent := range fleetServer.Components {
		slug := fleetComponent.Name
		alloyComponents, ok := alloyComponentsMap[slug]
		if !ok {
			match = false
			// no slug in alloy list, fleet component not in alloy list
			fields := []zap.Field{
				zap.String("device.id", fleetServer.ID),
				zap.String("component", slug),
				zap.String("expected(alloy)", "nil"),
				zap.String("current(fleetdb)", fleetComponent.Firmware.Installed),
			}
			log.Warn("fleetdb component not listed in alloy", fields...)
			continue
		}

		var listedByAlloy bool
		for _, alloyComponent := range alloyComponents {
			if alloyComponent.Vendor+alloyComponent.Model == fleetComponent.Vendor+fleetComponent.Model {
				// fleet component in alloy list
				listedByAlloy = true
				if alloyComponent.Firmware.Installed != fleetComponent.Firmware.Installed {
					match = false
					fields := []zap.Field{
						zap.String("device.id", fleetServer.ID),
						zap.String("component", slug),
						zap.String("expected(alloy)", alloyComponent.Firmware.Installed),
						zap.String("current(fleetdb)", fleetComponent.Firmware.Installed),
					}
					log.Warn("component firmware needs update", fields...)
				}
				break
			}
		}

		if !listedByAlloy {
			// alloy did not report hardware that has previously been in fleetdb
			match = false
			fields := []zap.Field{
				zap.String("device.id", fleetServer.ID),
				zap.String("component", slug),
				zap.String("expected(alloy)", "nil"),
				zap.String("current(fleetdb)", fleetComponent.Firmware.Installed),
			}
			log.Warn("fleetdb component not listed in alloy", fields...)
		}
	}
	return match
}
