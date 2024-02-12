package controllers

import (
	"context"
	"fmt"
	database "ideanest/pkg"
	model "ideanest/pkg/database/mongodb/models"
	"ideanest/pkg/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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
	var user model.User

	db := database.GetDB().Collection("users")

	err := c.BindJSON(&user)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var dbResult bson.M

	err = db.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&dbResult)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "This user doesn't exist",
			})
			return
		} else {
			log.Printf("Error querying database: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	isPasswordsMatch := utils.CheckPasswordHash(dbResult["password"].(string), user.Password)

	if !isPasswordsMatch {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid password",
		})
		return
	}

	tokenDetails, err := utils.CreateToken(dbResult["_id"].(primitive.ObjectID).Hex())

	if err != nil {
		log.Fatal("Error creating token: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = utils.CreateAuthWithRefresh(dbResult["_id"].(primitive.ObjectID).Hex(), tokenDetails)

	if err != nil {
		log.Fatal("Error creating auth: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(200, gin.H{
		"message":       "User signed in successfully",
		"access_token":  tokenDetails.AccessToken,
		"refresh_token": tokenDetails.RefreshToken,
	})
	return
}

func RefreshToken(c *gin.Context) {
	var tokenDetails bson.M

	err := c.BindJSON(&tokenDetails)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	refresh_token, ok := tokenDetails["refresh_token"]

	if !ok || refresh_token == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Refresh token is required",
		})
		return
	}

	refresh_token, ok = tokenDetails["refresh_token"].(string)

	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Refresh token is required",
		})
		return
	}

	if refresh_token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Refresh token is required",
		})
		return
	}

	// Check if refresh token exists in Redis

	_, err = database.GetRedis().Get(context.Background(), "r__"+refresh_token.(string)).Result()

	if err != nil {
		if err == database.RedisError() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Refresh token is expired or doesn't exist",
			})
			return
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	// Check if refresh token is not expired
	tokenData, err := utils.ExtractTokenMetadata(refresh_token.(string))

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Refresh token is expired or doesn't exist",
		})
		return
	}

	// Get user ID from refresh token
	userId, ok := tokenData.Claims.(jwt.MapClaims)["user_id"].(string)

	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Generate new token
	newToken, err := utils.GenAccessToken(userId)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Save new token in Redis
	err = utils.CreateAuth(userId, newToken)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(200, gin.H{
		"message":       "Token refreshed successfully",
		"access_token":  newToken.AccessToken,
		"refresh_token": refresh_token,
	})
	return
}

func RevokeRefreshToken(c *gin.Context) {
	var tokenDetails bson.M

	err := c.BindJSON(&tokenDetails)

	if err != nil {
		fmt.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	refresh_token, ok := tokenDetails["refresh_token"]

	if !ok || refresh_token == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Refresh token is required",
		})
		return
	}

	refresh_token, ok = tokenDetails["refresh_token"].(string)

	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Refresh token is required",
		})
		return
	}

	if refresh_token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Refresh token is required",
		})
		return
	}

	// Check if refresh token exists in Redis

	_, err = database.GetRedis().Get(context.Background(), "r__"+refresh_token.(string)).Result()

	if err != nil {
		if err == database.RedisError() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Refresh token is expired or doesn't exist",
			})
			return
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	// Delete refresh token from Redis
	_, err = database.GetRedis().Del(context.Background(), "r__"+refresh_token.(string)).Result()

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(200, gin.H{
		"message": "Refresh token revoked successfully",
	})
	return
}
