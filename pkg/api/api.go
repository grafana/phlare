// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/cortexproject/cortex/blob/master/pkg/api/api.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/felixge/fgprof"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/dskit/kv/memberlist"
	agentv1 "github.com/grafana/phlare/api/gen/proto/go/agent/v1"
	"github.com/grafana/phlare/api/gen/proto/go/agent/v1/agentv1connect"
	"github.com/grafana/phlare/api/gen/proto/go/push/v1/pushv1connect"
	"github.com/grafana/phlare/api/gen/proto/go/querier/v1/querierv1connect"
	statusv1 "github.com/grafana/phlare/api/gen/proto/go/status/v1"
	"github.com/grafana/phlare/api/openapiv2"
	"github.com/grafana/phlare/pkg/agent"
	"github.com/grafana/phlare/pkg/distributor"
	"github.com/grafana/phlare/pkg/ingester/pyroscope"
	"github.com/grafana/phlare/pkg/querier"
	"github.com/grafana/phlare/pkg/util"
	"github.com/grafana/phlare/pkg/util/gziphandler"
	"github.com/grafana/phlare/pkg/validation/exporter"
	"github.com/grafana/phlare/public"
	grpcgw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/common/middleware"
	"github.com/weaveworks/common/server"
)

type Config struct {
	// The following configs are injected by the upstream caller.
	ServerPrefix       string               `yaml:"-"`
	HTTPAuthMiddleware middleware.Interface `yaml:"-"`
}

type API struct {
	AuthMiddleware middleware.Interface

	cfg       Config
	server    *server.Server
	logger    log.Logger
	indexPage *IndexPageContent

	grpcGatewayMux *grpcgw.ServeMux
	auth           connect.Option
}

func New(cfg Config, s *server.Server, grpcGatewayMux *grpcgw.ServeMux, auth connect.Option, logger log.Logger) (*API, error) {
	// Ensure the encoded path is used. Required for the rules API
	s.HTTP.UseEncodedPath()

	api := &API{
		cfg:            cfg,
		AuthMiddleware: cfg.HTTPAuthMiddleware,
		server:         s,
		logger:         logger,
		indexPage:      NewIndexPageContent(),
		grpcGatewayMux: grpcGatewayMux,
		auth:           auth,
	}

	// If no authentication middleware is present in the config, use the default authentication middleware.
	if cfg.HTTPAuthMiddleware == nil {
		api.AuthMiddleware = middleware.AuthenticateUser
	}

	return api, nil
}

// RegisterRoute registers a single route enforcing HTTP methods. A single
// route is expected to be specific about which HTTP methods are supported.
func (a *API) RegisterRoute(path string, handler http.Handler, auth, gzipEnabled bool, method string, methods ...string) {
	methods = append([]string{method}, methods...)
	level.Debug(a.logger).Log("msg", "api: registering route", "methods", strings.Join(methods, ","), "path", path, "auth", auth, "gzip", gzipEnabled)
	a.newRoute(path, handler, false, auth, gzipEnabled, methods...)
}

func (a *API) RegisterRoutesWithPrefix(prefix string, handler http.Handler, auth, gzipEnabled bool, methods ...string) {
	level.Debug(a.logger).Log("msg", "api: registering route", "methods", strings.Join(methods, ","), "prefix", prefix, "auth", auth, "gzip", gzipEnabled)
	a.newRoute(prefix, handler, true, auth, gzipEnabled, methods...)
}

func (a *API) newRoute(path string, handler http.Handler, isPrefix, auth, gzip bool, methods ...string) (route *mux.Route) {
	if auth {
		handler = a.AuthMiddleware.Wrap(handler)
	}
	if gzip {
		handler = gziphandler.GzipHandler(handler)
	}
	if isPrefix {
		route = a.server.HTTP.PathPrefix(path)
	} else {
		route = a.server.HTTP.Path(path)
	}
	if len(methods) > 0 {
		route = route.Methods(methods...)
	}
	route = route.Handler(handler)

	return route
}

// RegisterAPI registers the standard endpoints associated with a running Mimir.
func (a *API) RegisterAPI(statusService statusv1.StatusServiceServer) error {
	// register index page
	a.RegisterRoute("/", indexHandler("", a.indexPage), false, true, "GET")
	// register grpc-gateway api
	a.RegisterRoutesWithPrefix("/api", a.grpcGatewayMux, false, true, "GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS")
	// expose openapiv2 definition
	openapiv2Handler, err := openapiv2.Handler()
	if err != nil {
		return fmt.Errorf("unable to initialize openapiv2 handler: %w", err)
	}
	a.RegisterRoute("/api/openapiv2", openapiv2Handler, false, true, "GET")
	a.indexPage.AddLinks(openAPIDefinitionWeight, "OpenAPI definition", []IndexPageLink{
		{Desc: "Swagger JSON", Path: "/api/swagger.json"},
	})
	// register fgprof
	a.RegisterRoute("/debug/fgprof", fgprof.Handler(), false, true, "GET")
	// register static assets
	a.RegisterRoutesWithPrefix("/static/", http.FileServer(http.FS(staticFiles)), false, true, "GET")
	// register ui
	uiAssets, err := public.Assets()
	if err != nil {
		return fmt.Errorf("unable to initialize the ui: %w", err)
	}
	a.RegisterRoutesWithPrefix("/ui/", http.FileServer(uiAssets), false, true, "GET")
	// register status service providing config and buildinfo at grpc gateway
	if err := statusv1.RegisterStatusServiceHandlerServer(context.Background(), a.grpcGatewayMux, statusService); err != nil {
		return err
	}
	a.indexPage.AddLinks(buildInfoWeight, "Build information", []IndexPageLink{
		{Desc: "Build information", Path: "/api/v1/status/buildinfo"},
	})
	a.indexPage.AddLinks(configWeight, "Current config", []IndexPageLink{
		{Desc: "Including the default values", Path: "/api/v1/status/config"},
		{Desc: "Only values that differ from the defaults", Path: "/api/v1/status/config/diff"},
		{Desc: "Default values", Path: "/api/v1/status/config/default"},
	})
	return nil
}

// RegisterRuntimeConfig registers the endpoints associates with the runtime configuration
func (a *API) RegisterRuntimeConfig(runtimeConfigHandler http.HandlerFunc, userLimitsHandler http.HandlerFunc) {
	a.RegisterRoute("/runtime_config", runtimeConfigHandler, false, true, "GET")
	a.RegisterRoute("/api/v1/tenant_limits", userLimitsHandler, true, true, "GET")
	a.indexPage.AddLinks(runtimeConfigWeight, "Current runtime config", []IndexPageLink{
		{Desc: "Entire runtime config (including overrides)", Path: "/runtime_config"},
		{Desc: "Only values that differ from the defaults", Path: "/runtime_config?mode=diff"},
	})
}

// RegisterOverridesExporter registers the endpoints associated with the overrides exporter.
func (a *API) RegisterOverridesExporter(oe *exporter.OverridesExporter) {
	a.RegisterRoute("/overrides-exporter/ring", http.HandlerFunc(oe.RingHandler), false, true, "GET", "POST")
	a.indexPage.AddLinks(defaultWeight, "Overrides-exporter", []IndexPageLink{
		{Desc: "Ring status", Path: "/overrides-exporter/ring"},
	})
}

// RegisterDistributor registers the endpoints associated with the distributor.
func (a *API) RegisterDistributor(d *distributor.Distributor, multitenancyEnabled bool) {
	a.server.HTTP.Handle("/pyroscope/ingest", util.AuthenticateUser(multitenancyEnabled).Wrap(pyroscope.NewPyroscopeIngestHandler(d, a.logger)))
	pushv1connect.RegisterPusherServiceHandler(a.server.HTTP, d, a.auth)
	a.RegisterRoute("/distributor/ring", d, false, true, "GET", "POST")
	a.indexPage.AddLinks(defaultWeight, "Distributor", []IndexPageLink{
		{Desc: "Ring status", Path: "/distributor/ring"},
	})
}

func (a *API) RegisterMemberlistKV(pathPrefix string, kvs *memberlist.KVInitService) {
	a.RegisterRoute("/memberlist", MemberlistStatusHandler(pathPrefix, kvs), false, true, "GET")
	a.indexPage.AddLinks(memberlistWeight, "Memberlist", []IndexPageLink{
		{Desc: "Status", Path: "/memberlist"},
	})
}

// RegisterRing registers the ring UI page associated with the distributor for writes.
func (a *API) RegisterRing(r http.Handler) {
	a.RegisterRoute("/ring", r, false, true, "GET", "POST")
	a.indexPage.AddLinks(defaultWeight, "Ingester", []IndexPageLink{
		{Desc: "Ring status", Path: "/ring"},
	})
}

// RegisterQuerier registers the endpoints associated with the querier.
func (a *API) RegisterQuerier(svc querierv1connect.QuerierServiceHandler, multitenancyEnabled bool) {
	handlers := querier.NewHTTPHandlers(svc)
	querierv1connect.RegisterQuerierServiceHandler(a.server.HTTP, svc, a.auth)

	a.server.HTTP.Handle("/pyroscope/render", util.AuthenticateUser(multitenancyEnabled).Wrap(http.HandlerFunc(handlers.Render)))
	a.server.HTTP.Handle("/pyroscope/label-values", util.AuthenticateUser(multitenancyEnabled).Wrap(http.HandlerFunc(handlers.LabelValues)))
}

func (a *API) RegisterAgent(ag *agent.Agent) error {
	// register endpoint at grpc gateway
	if err := agentv1.RegisterAgentServiceHandlerServer(context.Background(), a.grpcGatewayMux, ag); err != nil {
		return err
	}
	agentv1connect.RegisterAgentServiceHandler(a.server.HTTP, ag.ConnectHandler())

	return nil
}
