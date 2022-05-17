package Structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Structure used to access all fields of a user in the database
type Users struct {
	//ID            primitive.ObjectID `bson:"-" json:"id"`
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username      string             `json:"username"`
	Discriminator string             `json:"discriminator"`
	Avatar        string             `json:"avatar"`
	Bot           bool               `json:"bot"`
	System        bool               `json:"system"`
	MFA           bool               `json:"mfa"`
	Banner        string             `json:"banner"`
	Accent_Color  int                `json:"accent_color"`
	Locale        string             `json:"locale"`
	Email         string             `json:"email"`
	EmailVerified bool               `json:"emailverified"`
	Password      []byte             `json:"password,omitempty"`
	Role          Role               `json:"role"`
	RegisterTime  int64              `json:"registertime"`
}
type Role struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Expiry   int64  `json:"expiry"`
	Creation int64  `json:"creation"`
}
