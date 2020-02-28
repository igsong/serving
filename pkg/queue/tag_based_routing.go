package queue

import (
	"context"
	"fmt"
	"net/http"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"

	pkgmetrics "knative.dev/pkg/metrics"
	"knative.dev/serving/pkg/metrics"
	"knative.dev/serving/pkg/network"
)

var (
	requestWithInvalidTagHeaderCountM = stats.Int64(
		"request_with_invalid_tag_header_count",
		"The number of requests with the tag header which is not matched with the tag reference header",
		stats.UnitDimensionless)
)

func getUniqueHeader(r *http.Request, headerName string) (string, error) {
	// Check if there are multiple header entries
	if len(r.Header[headerName]) > 1 {
		return "", fmt.Errorf("multiple %s header entries are not permitted", headerName)
	}

	return r.Header.Get(headerName), nil
}

// NewTagBasedRoutingHandler create a handler for detecting inconsistency between the tag header coming with a request and the reference tag header denoting the route defined in Ingress
func NewTagBasedRoutingHandler(next http.Handler, ns string, service string, config string, rev string, enableFallback bool) http.Handler {
	keys := append(metrics.CommonRevisionKeys,
		metrics.TagActualKey,
		metrics.TagExpectedKey)

	var statsCtx context.Context

	if err := view.Register(&view.View{
		Description: requestWithInvalidTagHeaderCountM.Description(),
		Measure:     requestWithInvalidTagHeaderCountM,
		Aggregation: view.Count(),
		TagKeys:     keys,
	}); err != nil {
		statsCtx = nil
	}
	statsCtx, err := metrics.RevisionContext(ns, service, config, rev)
	if err != nil {
		statsCtx = nil
	}
	// To prevent use of appended
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// uniqueness check
		tag, err := getUniqueHeader(r, network.TagHeaderName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		tagRef, err := getUniqueHeader(r, network.TagRefHeaderName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer func() {
			w.Header().Add(network.TagRefHeaderName, tagRef)
		}()

		if !enableFallback && len(tag) > 0 && tag != tagRef {
			if statsCtx != nil {
				ctx := metrics.AugmentWithActualAndExpectedTagName(statsCtx, GetTagRefName(r), GetTagName(r))
				pkgmetrics.RecordBatch(ctx, requestWithInvalidTagHeaderCountM.M(1))
			}

			if tagRef == network.DefaultTargetHeaderValue {
				// If a request has different values on tag and tagrRef, it is an invalid request.
				// Since such case happen when a user make a request with non-existing tag, here, NotFound is returned.
				http.Error(w, "tag not found", http.StatusNotFound)
			} else {
				http.Error(w, "inconsistent tag is provided", http.StatusBadGateway)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}
