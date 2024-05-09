package routes

import (
	rivets "github.com/metal-toolbox/rivets/types"
	"go.uber.org/zap"
)

// compareComponents compares components between two rivets.Server.
// It logs differences and return false if there are differences.
func compareComponents(fleetServer, alloyServer *rivets.Server, log *zap.Logger) {
	alloyMap := componentsToMap(alloyServer.Components)
	fleetMap := componentsToMap(fleetServer.Components)
	log.Debug("enumerating incoming")
	for k, v := range alloyMap {
		log.With(
			zap.String("component.name", k),
			zap.Int("component.count", len(v)),
		).Debug("incoming component")
	}
	log.Debug("enumerating existing")
	for k, v := range fleetMap {
		log.With(
			zap.String("component.name", k),
			zap.Int("component.count", len(v)),
		).Debug("existing component")
	}
}

type componentMap map[string][]*rivets.Component

func componentsToMap(cs []*rivets.Component) componentMap {
	theMap := make(map[string][]*rivets.Component)
	for _, c := range cs {
		name := c.Name
		// cSlice can be nil. Appending to nil is OK.
		cSlice := theMap[name]
		cSlice = append(cSlice, c)
		theMap[name] = cSlice
	}
	return theMap
}
