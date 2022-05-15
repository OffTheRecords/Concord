package Authentication

import (
	"Concord/CustomErrors"
	"Concord/Structures"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"reflect"
	"regexp"
	"time"
)

//Will check if the entered struct has any uninitialized types (either nil or empty string)
func FieldValidityCheck(value reflect.Value, ref reflect.Type) CustomErrors.GenericErrors {
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Interface() == nil || value.Field(i).Interface() == "" {
			return CustomErrors.NewGenericError(4006, "empty fields")
		}
		if ref.Field(i).Name == "Email" {
			val := value.Field(i).String()
			argMatched, err := regexp.MatchString("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", val)
			if argMatched == false || err != nil {
				return CustomErrors.NewGenericError(4001, "invalid email format")
			}
		}
		if ref.Field(i).Name == "Username" {
			val := value.Field(i).String()
			argMatched, err := regexp.MatchString("\\A[a-zA-Z0-9_-]{2,32}$", val)
			if argMatched == false || err != nil {
				return CustomErrors.NewGenericError(4002, "invalid username format")
			}

		}
		if ref.Field(i).Name == "Password" {
			val := value.Field(i).String()
			argMatched, err := regexp.MatchString("\\A.{8,128}$", val)
			if argMatched == false || err != nil {
				return CustomErrors.NewGenericError(4003, "invalid password format")
			}

		}
	}
	return nil
}

func FieldEmptyCheck(value reflect.Value) CustomErrors.GenericErrors {
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Interface() == nil || value.Field(i).Interface() == "" {
			return CustomErrors.NewGenericError(4006, "empty fields")
		}
	}
	return nil
}

func makeRelativeDir(path string) CustomErrors.GenericErrors {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, 0750)
		if err != nil {
			return CustomErrors.NewGenericError(5005, err.Error())
		}
	}
	return nil
}

func GetUserFromDB(email string, dbClient *mongo.Database) (Structures.Users, CustomErrors.GenericErrors) {
	var user Structures.Users
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := dbClient.Collection(GetAuthCollection()).FindOne(ctx, bson.D{{"email", email}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, CustomErrors.NewGenericError(4010, "No matching email")
		} else {
			return user, CustomErrors.NewGenericError(5014, err.Error())
		}
	}
	return user, nil
}

func GetUserUsingIDFromDB(userID string, dbClient *mongo.Database) (Structures.Users, CustomErrors.GenericErrors) {
	var user Structures.Users

	userHex, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return user, CustomErrors.NewGenericError(5019, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = dbClient.Collection(GetAuthCollection()).FindOne(ctx, bson.D{{"_id", userHex}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, CustomErrors.NewGenericError(4013, "No matching _id")
		} else {
			return user, CustomErrors.NewGenericError(5014, err.Error())
		}
	}
	return user, nil
}
