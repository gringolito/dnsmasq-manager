package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gringolito/dnsmasq-manager/api/middleware/fiberswagger"
)

type Router interface {
	AddMetricsRoute(cfg monitor.Config)
	AddSwaggerUIRoute(openApiSpecFile string)
	AddApiV1Route(prefix string, routes func(fiber.Router), name ...string)
	AuthenticationHandler(roles ...string) fiber.Handler
}

type router struct {
	root  fiber.Router
	api   fiber.Router
	apiv1 fiber.Router
	mw    Middleware
}

func NewRouter(root fiber.Router, mw Middleware) Router {
	root.Use(mw.Recovery())
	root.Use(mw.Logger())

	api := root.Group("/api")
	api.Use(mw.RequestId())

	apiv1 := api.Group("/v1")

	return &router{
		root:  root,
		api:   api,
		apiv1: apiv1,
		mw:    mw,
	}
}

func (r *router) AddMetricsRoute(cfg monitor.Config) {
	r.root.Get("/metrics", monitor.New(cfg))
}

func (r *router) AddSwaggerUIRoute(openApiSpecFile string) {
	fiberswagger.Router(r.root, fiberswagger.Config{
		BasePath: "/openapi",
		FilePath: openApiSpecFile,
	})
}

func (r *router) AddApiV1Route(prefix string, routes func(fiber.Router), name ...string) {
	r.apiv1.Route(prefix, routes, name...)
}

func (r *router) AuthenticationHandler(roles ...string) fiber.Handler {
	return r.mw.Authentication(roles...)
}
