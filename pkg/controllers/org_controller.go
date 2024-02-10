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
			"message": "An organization with this name already exists. Please choose a different name.",
		})
		return
	}

	organization.ID = primitive.NewObjectID()

	_, err = db.InsertOne(context.TODO(), organization)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"organization_id": organization.ID,
	})
}

func GetAllOrganizations(c *gin.Context) {
	var organizations []bson.M
	cursor, err := database.GetDB().Collection("organizations").Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = cursor.All(context.TODO(), &organizations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, organizations)
}

func GetOrganizationById(c *gin.Context) {
	var organization model.Organization
	id := c.Param("id")

	db := database.GetDB().Collection("organizations")

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = db.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&organization)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else {
			log.Printf("Error querying database: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, organization)
	return
}

func UpdateOrganization(c *gin.Context) {
	var result bson.M
	var organization model.Organization
	id := c.Param("id")
	db := database.GetDB().Collection("organizations")
	err := c.BindJSON(&organization)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	oid, err := primitive.ObjectIDFromHex(id)

	err = db.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Organization not found",
			})
			return
		} else {
			log.Printf("Error querying database: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	err = db.FindOne(context.TODO(), bson.M{"name": organization.Name}).Decode(&result)

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
			"message": "An organization with this name already exists. Please choose a different name.",
		})
		return
	}

	_, err = db.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.M{"$set": organization})

	c.JSON(200, gin.H{
		"organization_id": id,
		"name":            organization.Name,
		"description":     organization.Description,
	})
}

func DeleteOrganization(c *gin.Context) {
	db := database.GetDB().Collection("organizations")
	organizationId := c.Param("id")

	oid, err := primitive.ObjectIDFromHex(organizationId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, err := db.DeleteOne(context.TODO(), bson.M{"_id": oid})

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if result.DeletedCount < 1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Organization deleted successfully",
	})
}

func InviteUserToOrganization(c *gin.Context) {
	organizationId := c.Param("id")

	oid, err := primitive.ObjectIDFromHex(organizationId)

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var body bson.M
	var organization model.Organization
	var orgUsers model.OrganizationMember
	var user model.User

	err = c.BindJSON(&body)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	email := body["email"].(string)

	db := database.GetDB().Collection("organizations")

	err = db.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&organization)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Organization not found",
			})
			return
		} else {
			log.Printf("Error querying database: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	err = database.GetDB().Collection("users").FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "User not found",
			})
			return
		} else {
			log.Printf("Error querying database: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	// Check if user is already a member of the organization
	var result bson.M
	err = database.GetDB().Collection("organization_members").FindOne(context.TODO(), bson.M{"user_id": user.ID, "org_id": organization.ID}).Decode(&result)

	if err == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "User is already a member of this organization",
		})
		return
	}

	// Create a new organization member

	orgUsers.ID = primitive.NewObjectID()
	orgUsers.UserID = user.ID
	orgUsers.AccessLevel = "r"
	orgUsers.OrgID = organization.ID

	_, err = database.GetDB().Collection("organization_members").InsertOne(context.TODO(), orgUsers)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Add the user to the organization

	_, err = database.GetDB().Collection("organizations").UpdateOne(
		context.TODO(),
		bson.M{"_id": oid},
		bson.M{
			"$push": bson.M{"organization_members": orgUsers.ID}, // Add the new member.
		},
	)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation sent successfully to " + email + " for organization " + organization.Name,
	})
}
