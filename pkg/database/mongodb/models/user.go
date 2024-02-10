package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string             `bson:"name" json:"name"`
	Email    string             `bson:"email,unique,omitempty" json:"email,unique,omitempty"`
	Password string             `bson:"password,omitempty" json:"password,omitempty"`
}
