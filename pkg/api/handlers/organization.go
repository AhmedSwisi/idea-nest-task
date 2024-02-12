package handlers

import (
	"ideanest/pkg/controllers"

	"github.com/gin-gonic/gin"
)

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
