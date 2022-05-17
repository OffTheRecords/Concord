package Authentication

import (
	"Concord/CustomErrors"
	"Concord/Structures"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

func GetAuthCollection() string {
	return "Users"
}

func RegisterUser(email string, username string, password string, dbClient *mongo.Database) (primitive.ObjectID, CustomErrors.GenericErrors) {
	ciphertext, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		CustomErrors.LogError(4011, "WARNING", false, err)
		return primitive.NilObjectID, CustomErrors.NewGenericError(4011, "Password encryption error")
	}

	//Update time of unverified role
	Structures.UNVERIFIED_ROLE.Creation = time.Now().Unix()
	registerDB := Structures.Users{Email: email, Username: username, Password: ciphertext, EmailVerified: false, RegisterTime: time.Now().Unix(), Role: Structures.UNVERIFIED_ROLE}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := dbClient.Collection(GetAuthCollection()).InsertOne(ctx, registerDB)
	if err != nil {
		if strings.Contains(err.Error(), "dup key") {
			fmt.Println(err.Error())
			return primitive.NilObjectID, CustomErrors.NewGenericError(4007, "registration failed, email address taken")
		} else {
			CustomErrors.LogError(5016, "WARNING", false, err)
			return primitive.NilObjectID, CustomErrors.NewGenericError(5016, err.Error())
		}
	}

	return res.InsertedID.(primitive.ObjectID), nil
}
