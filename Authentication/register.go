package Authentication

import (
	"Concord/CustomErrors"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

//TODO Need to generate user ID's to differentiate users.
type RegisterUserDB struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email-verified"`
	Username      string `json:"username"`
	Password      []byte `json:"password"`
}

func getAuthCollection() string {
	return "Users"
}

func RegisterUser(email string, username string, password string, dbClient *mongo.Database) CustomErrors.GenericErrors {
	ciphertext, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		CustomErrors.LogError(4009, "WARNING", false, err)
		return CustomErrors.NewGenericError(4009, "Password encryption error")
	}

	registerDB := RegisterUserDB{Email: email, Username: username, Password: ciphertext, EmailVerified: false}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = dbClient.Collection(getAuthCollection()).InsertOne(ctx, registerDB)
	if err != nil {
		if strings.Contains(err.Error(), "dup key") {
			return CustomErrors.NewGenericError(4007, "registration failed, email address taken")
		} else {
			CustomErrors.LogError(4008, "WARNING", false, err)
			return CustomErrors.NewGenericError(4008, "registration failed")
		}

	}
	return nil
}
