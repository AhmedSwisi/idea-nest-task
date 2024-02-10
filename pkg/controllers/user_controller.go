package controllers

import (
	"context"
	database "ideanest/pkg"
	model "ideanest/pkg/database/mongodb/models"
	"ideanest/pkg/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SignUp(c *gin.Context) {
	var user model.User

	db := database.GetDB().Collection("users")
	err := c.BindJSON(&user)

	if err != nil {
		log.Fatal("1")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Check if user exists before
	var dbResult bson.M

	err = db.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&dbResult)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// It's fine because it means that there's no user with this email
		} else {
			log.Printf("Error querying database: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "This user exists",
		})
		return
	}

	user.ID = primitive.NewObjectID()
	user.Password, err = utils.HashPassword(user.Password)

	if err != nil {
		log.Fatal("4")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	_, err = db.InsertOne(context.TODO(), user)

	if err != nil {
		log.Fatal("5")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
	})
	return
}

func SignIn(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "User signed in successfully",
	})
	return
}

func RefreshToken(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "RefreshToken",
	})
}
