package Structures

import "go.mongodb.org/mongo-driver/bson/primitive"

//Structure used to access all fields of a user in the database
type Users struct {
	ID            primitive.ObjectID `bson:"-" json:"id"`
	Username      string             `json:"username"`
	Email         string             `json:"email"`
	EmailVerified bool               `json:"emailverified"`
	Password      []byte             `json:"password"`
	Role          Role               `json:"roles,omitempty"`
	RegisterTime  int64              `json:"registertime"`
}
type Role struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Expiry   int64  `json:"expiry"`
	Creation int64  `json:"creation"`
}
