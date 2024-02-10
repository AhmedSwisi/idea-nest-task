package handlers

import (
	"ideanest/pkg/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("/signup", controllers.SignUp)
	router.POST("/signin", controllers.SignIn)
	router.POST("/refresh-token", controllers.RefreshToken)
}
