package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type OrganizationMember struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	AccessLevel string             `bson:"access_level" json:"access_level,omitempty"`
	OrgID       primitive.ObjectID `bson:"org_id,omitempty" json:"org_id,omitempty"`
}

type Organization struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name,omitempty" json:"name,omitempty"`
	Description string             `bson:"description" json:"description"`
}
