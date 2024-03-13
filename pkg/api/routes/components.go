package routes

import (
	"context"
	"time"

	"github.com/google/uuid"
	fleetdb "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	rdb "github.com/metal-toolbox/rivets/fleetdb"
	rivets "github.com/metal-toolbox/rivets/types"
	"go.uber.org/zap"
)

var fleetDBTimeout = 3 * time.Minute

// this is a map of "component_type_name" to the actual inventory data for each component
type serverComponents map[string][]*rivets.Component

func fetchServerComponents(client *fleetdb.Client, srvid uuid.UUID, l *zap.Logger) (serverComponents, error) {
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
