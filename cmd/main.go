package main

import (
	"fmt"
	database "ideanest/pkg"
	handlers "ideanest/pkg/api/handlers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	database.Init(os.Getenv("MONGO_URI"), "development")

	fmt.Println("Connected to MongoDB")

	defer func() {
		err := database.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello world!",
		})
	})

	orgRoutes := router.Group("/")

	handlers.AuthRoutes(router)
	handlers.OrganizationRoutes(orgRoutes)

	router.Run(":8080")
}