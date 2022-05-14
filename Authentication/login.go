package Authentication

import (
	"Concord/CustomErrors"
	"Concord/Structures"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func Login(email string, password string, dbClient *mongo.Database) (Structures.AuthTokens, CustomErrors.GenericErrors) {
	user, gerr := GetUserFromDB(email, dbClient)
	jwtToken := Structures.AuthTokens{}
	if gerr != nil {
		return jwtToken, gerr
	}

	err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return jwtToken, CustomErrors.NewGenericError(4009, "Invalid password")
	}

	jwtToken, err = GenerateJWT(user)
	if err != nil {
		CustomErrors.LogError(5015, CustomErrors.LOG_WARNING, false, err)
		return jwtToken, CustomErrors.NewGenericError(5015, err.Error())
	}

	return jwtToken, nil
}
