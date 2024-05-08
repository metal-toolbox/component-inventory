package routes

import (
	rivets "github.com/metal-toolbox/rivets/types"
	"go.uber.org/zap"
)

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
		_, ok := alloyComponentsMap[slug]
		if !ok {
			match = false
			// no slug in alloy list, fleet component not in alloy list
			fields := []zap.Field{
				zap.String("device.id", fleetServer.ID),
				zap.String("component", slug),
			}
			log.Warn("fleetdb component not listed in alloy", fields...)
			continue
		}

		log.With(
			zap.String("device.id", fleetServer.ID),
			zap.String("component", slug),
		).Info("placeholder for server component firmware check")
	}
	return match
}
