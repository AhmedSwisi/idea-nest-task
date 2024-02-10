package handlers

import (
	"ideanest/pkg/controllers"

	"github.com/gin-gonic/gin"
)

// r.GET("/organization", handler.GetOrganizations)
// r.GET("/organization/:id", handler.GetOrganizationById)
// r.POST("organization/:id/invite", handler.InviteUserToOrganization)
// r.POST("/organization", handler.AddOrganization)
// r.PUT("organization/:id", handler.UpdateOrganization)
// r.DELETE("organization/:id", handler.DeleteOrganization)

func OrganizationRoutes(router *gin.RouterGroup) {
	organizationRouter := router.Group("/organization")
	{
		organizationRouter.GET("/:id", controllers.GetOrganizationById)
		organizationRouter.GET("/", controllers.GetAllOrganizations)
		organizationRouter.POST("/", controllers.AddOrganization)
		organizationRouter.POST("/:id/invite", controllers.InviteUserToOrganization)
		organizationRouter.PUT("/:id", controllers.UpdateOrganization)
		organizationRouter.DELETE("/:id", controllers.DeleteOrganization)
	}

}
