package Structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Structure used to access all fields of a user in the database
type Users struct {
	//ID            primitive.ObjectID `bson:"-" json:"id"`
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username      string             `json:"username"`
	Discriminator string             `json:"discriminator,omitempty"`
	Avatar        string             `json:"avatar,omitempty"`
	Bot           bool               `json:"bot,omitempty"`
	System        bool               `json:"system,omitempty"`
	MFA           bool               `json:"mfa,omitempty"`
	Banner        string             `json:"banner,omitempty"`
	Accent_Color  int                `json:"accent_color,omitempty"`
	Locale        string             `json:"locale,omitempty"`
	Email         string             `json:"email,omitempty"`
	EmailVerified bool               `json:"emailverified,omitempty"`
	Password      []byte             `json:"password,omitempty"`
	Role          Role               `json:"role,omitempty"`
	RegisterTime  int64              `json:"registertime,omitempty"`
}
type Role struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Expiry   int64  `json:"expiry,omitempty"`
	Creation int64  `json:"creation,omitempty"`
}
