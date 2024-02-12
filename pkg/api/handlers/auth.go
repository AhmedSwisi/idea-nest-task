package handlers

import (
	middlewares "ideanest/pkg/api/middleware"
	"ideanest/pkg/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("/signup", controllers.SignUp)
	router.POST("/signin", controllers.SignIn)
	router.POST("/refresh-token", controllers.RefreshToken)
	router.Use(middlewares.JWTMiddleware).POST("/revoke-refresh-token", controllers.RevokeRefreshToken)
}
