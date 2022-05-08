package Authentication

import (
	"Concord/CustomErrors"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"reflect"
	"time"
)

type userDetails struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
	Role     string `json:"role"`
}

func Login(email string, password string, dbClient *mongo.Database) (string, error, time.Time) {
	opts := options.FindOne().SetSort(bson.D{{"age", 1}})
	var result userDetails
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := dbClient.Collection(getAuthCollection()).FindOne(ctx, bson.D{{"email", email}}, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "No matching email", err, time.Now()
		}
		log.Fatal(err)
	} else {
		err := bcrypt.CompareHashAndPassword(result.Password, []byte(password))

		if err != nil {
			CustomErrors.LogError(4009, "WARNING", false, err)
			return "Password encryption error", err, time.Now()
		}
		userDetails := userDetails{Email: email, Role: result.Role}
		jwtString, err, expiry := GenerateJWT(reflect.ValueOf(userDetails))
		return jwtString, err, expiry
	}
	return "No password match", nil, time.Now()
}
