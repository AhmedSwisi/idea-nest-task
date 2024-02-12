package utils

import (
	"context"
	"fmt"
	database "ideanest/pkg"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AtExpires    int64
	RtExpires    int64
}

func CreateToken(userId string) (*TokenDetails, error) {

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Hour * 24).Unix()
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()

	var err error

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = userId
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		return nil, err
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["user_id"] = userId
	rtClaims["exp"] = td.RtExpires

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		return nil, err
	}

	return td, nil
}

func CreateAuthWithRefresh(userId string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := database.GetRedis().Set(context.TODO(), "a__"+td.AccessToken, userId, at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}

	errRefresh := database.GetRedis().Set(context.TODO(), "r__"+td.RefreshToken, userId, rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}

	return nil
}

func GenAccessToken(userId string) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Hour * 24).Unix()

	var err error

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = userId
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		return nil, err
	}

	return td, nil

}

func CreateAuth(userId string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0)
	now := time.Now()

	errAccess := database.GetRedis().Set(context.TODO(), "a__"+td.AccessToken, userId, at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}

	return nil
}

func ExtractTokenMetadata(token string) (*jwt.Token, error) {
	tokenData, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	return tokenData, nil
}

// func ExtractTokenMetadata(token string) (*jwt.Token, error) {
// 	tokenData, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
// 		return []byte("JWT_SECRET"), nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	return tokenData, nil
// }
