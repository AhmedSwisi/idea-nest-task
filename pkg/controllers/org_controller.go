package controllers

import (
	"context"
	"fmt"
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

	// Check if the organization name is empty

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

type OrgMemberResponse struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	AccessLevel string `json:"access_level"`
}

type OrganizationResponse struct {
	OrganizationID string              `json:"organization_id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	OrgMembers     []OrgMemberResponse `json:"org_members"`
}

func GetAllOrganizations(c *gin.Context) {
	var response []OrganizationResponse
	var organizations []model.Organization
	var user model.User
	organizations_collection := database.GetDB().Collection("organizations")
	user_collection := database.GetDB().Collection("users")

	cursor, err := organizations_collection.Find(context.TODO(), bson.M{})

	userId, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
	}

	if err != nil {

		log.Printf("Error querying database: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Get the user from the database
	oid, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = user_collection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&user)

	if err != nil {
		log.Printf("Error querying database: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	fmt.Println("user fetched from context: ", user)

	if err = cursor.All(context.TODO(), &organizations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, org := range organizations {
		var orgResponse OrganizationResponse
		var orgMembersResponse []OrgMemberResponse

		oid := org.ID

		var orgMembers []model.OrganizationMember
		orgMembersCursor, err := database.GetDB().Collection("organization_members").Find(context.TODO(), bson.M{"org_id": oid})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err = orgMembersCursor.All(context.TODO(), &orgMembers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, member := range orgMembers {
			var user model.User
			err = user_collection.FindOne(context.TODO(), bson.M{"_id": member.UserID}).Decode(&user)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			orgMembersResponse = append(orgMembersResponse, OrgMemberResponse{
				Name:        user.Name,
				Email:       user.Email,
				AccessLevel: member.AccessLevel,
			})
		}

		orgResponse = OrganizationResponse{
			OrganizationID: oid.Hex(),
			Name:           org.Name,
			Description:    org.Description,
			OrgMembers:     orgMembersResponse,
		}

		response = append(response, orgResponse)
	}

	c.JSON(http.StatusOK, response)
}

func GetOrganizationById(c *gin.Context) {
	var response []OrganizationResponse
	var organizations model.Organization
	organizations_collection := database.GetDB().Collection("organizations")
	user_collection := database.GetDB().Collection("users")

	_id := c.Param("id")

	oid, err := primitive.ObjectIDFromHex(_id)

	// Check if the organization exists
	err = organizations_collection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&organizations)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	var orgResponse OrganizationResponse

	var orgMembersResponse []OrgMemberResponse

	var orgMembers []model.OrganizationMember

	orgMembersCursor, err := database.GetDB().Collection("organization_members").Find(context.TODO(), bson.M{"org_id": oid})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = orgMembersCursor.All(context.TODO(), &orgMembers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, member := range orgMembers {
		var user model.User
		err = user_collection.FindOne(context.TODO(), bson.M{"_id": member.UserID}).Decode(&user)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		orgMembersResponse = append(orgMembersResponse, OrgMemberResponse{
			Name:        user.Name,
			Email:       user.Email,
			AccessLevel: member.AccessLevel,
		})
	}

	orgResponse = OrganizationResponse{
		OrganizationID: oid.Hex(),
		Name:           organizations.Name,
		Description:    organizations.Description,
		OrgMembers:     orgMembersResponse,
	}

	response = append(response, orgResponse)
	c.JSON(http.StatusOK, response)
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

	if organization.Name != result["name"] {

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
	}

	value, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	uid, err := primitive.ObjectIDFromHex(value.(string))

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Check if the user is the owner of the organization
	err = database.GetDB().Collection("organization_members").FindOne(context.TODO(), bson.M{"org_id": oid, "user_id": uid}).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "You are not authorized to update this organization",
			})
			return
		} else {
			log.Printf("Error querying database: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	_, err = db.UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.M{"$set": organization})

	c.JSON(200, gin.H{
		"organization_id": id,
		"name":            organization.Name,
		"description":     organization.Description,
	})
}

func DeleteOrganization(c *gin.Context) {
	var result bson.M
	db := database.GetDB().Collection("organizations")
	organizationId := c.Param("id")

	oid, err := primitive.ObjectIDFromHex(organizationId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	value, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	uid, err := primitive.ObjectIDFromHex(value.(string))

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = database.GetDB().Collection("organization_members").FindOne(context.TODO(), bson.M{"org_id": oid, "user_id": uid}).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "You are not authorized to delete this organization",
			})
			return
		} else {
			log.Printf("Error querying database: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	dbResult, err := db.DeleteOne(context.TODO(), bson.M{"_id": oid})

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if dbResult.DeletedCount < 1 {
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
	orgUsers.AccessLevel = "read"
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
