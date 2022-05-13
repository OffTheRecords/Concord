package Authentication

import (
	"Concord/CustomErrors"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func Login(email string, password string, dbClient *mongo.Database) (string, CustomErrors.GenericErrors, time.Time) {
	user, gerr := GetUserFromDB(email, dbClient)
	if gerr != nil {
		return "", gerr, time.Now()
	}

	err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return "", CustomErrors.NewGenericError(4009, "Invalid password"), time.Now()
	}

	jwtString, err, expiry := GenerateJWT(user)
	if err != nil {
		CustomErrors.LogError(5015, CustomErrors.LOG_WARNING, false, err)
		return "", CustomErrors.NewGenericError(5015, err.Error()), time.Now()
	}

	return jwtString, nil, expiry
}
