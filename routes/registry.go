package routes

import (
	"field-service/clients"
	"field-service/controllers"
	fieldRoute "field-service/routes/field"
	fieldScheduleRoute "field-service/routes/field_schedule"
	timeRoute "field-service/routes/time"
	"github.com/gin-gonic/gin"
)

type Registry struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
	client     clients.IClientRegistry
}

type IRouterRegistry interface {
	Serve()
}

func NewRouteRegistry(group *gin.RouterGroup, controller controllers.IControllerRegistry, client clients.IClientRegistry) IRouterRegistry {
	return &Registry{
		group:      group,
		controller: controller,
		client:     client,
	}
}

func (r *Registry) Serve() {
	r.fieldRoute().Run()
	r.fieldScheduleRoute().Run()
	r.timeRoute().Run()
}

func (r *Registry) fieldRoute() fieldRoute.IFieldRoute {
	return fieldRoute.NewFieldRoute(r.group, r.controller, r.client)
}

func (r *Registry) fieldScheduleRoute() fieldScheduleRoute.IFieldScheduleRoute {
	return fieldScheduleRoute.NewFieldScheduleRoute(r.group, r.controller, r.client)
}

func (r *Registry) timeRoute() timeRoute.ITimeRoute {
	return timeRoute.NewTimeRoute(r.group, r.controller, r.client)
}
