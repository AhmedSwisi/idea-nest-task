package controllers

import (
	"context"
	database "ideanest/pkg"
	model "ideanest/pkg/database/mongodb/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddOrganization(c *gin.Context) {
	var organization model.Organization
	db := database.GetDB().Collection("organizations")
	err := c.BindJSON(&organization)

	if err != nil {
		log.Fatal("1")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var dbResult bson.M

	err = db.FindOne(context.TODO(), bson.M{"name": organization.Name}).Decode(&dbResult)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// no organization has this name
		} else {
			log.Printf("Error querying database: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "This organization already exists",
		})
		return
	}

	organization.ID = primitive.NewObjectID()

	_, err = db.InsertOne(context.TODO(), organization)

	if err != nil {
		log.Fatal("5")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"organization_id": organization.ID,
	})
}

func GetAllOrganizations(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Get all",
	})
}

func GetOrganizationById(c *gin.Context) {
	id := c.Param("id")
	for _, a := range organizations {

	}
	c.JSON(200, gin.H{
		"message": "Get by id",
	})
}

func UpdateOrganization(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "update org",
	})
}

func DeleteOrganization(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Delete",
	})
}

func InviteUserToOrganization(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Invite ",
	})
}
