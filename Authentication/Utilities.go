package Authentication

import (
	"errors"
	"reflect"
	"regexp"
)

//Will check if the entered struct has any uninitialized types (either nil or empty string)
func FieldValidityCheck(value reflect.Value, ref reflect.Type) error {
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Interface() == nil || value.Field(i).Interface() == "" {
			return errors.New("Empty fields")
		}
		if ref.Field(i).Name == "Email" {
			val := value.Field(i).String()
			argMatched, err := regexp.MatchString("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", val)
			if argMatched == false || err != nil {
				return errors.New("Incorrect email format")
			}
		}
		if ref.Field(i).Name == "Username" {
			val := value.Field(i).String()
			argMatched, err := regexp.MatchString("\\A[a-zA-Z0-9_-]{2,32}$", val)
			if argMatched == false || err != nil {
				return errors.New("Incorrect username format")
			}

		}
		if ref.Field(i).Name == "Password" {
			val := value.Field(i).String()
			argMatched, err := regexp.MatchString("\\A.{8,128}$", val)
			if argMatched == false || err != nil {
				return errors.New("Incorrect password format")
			}

		}
	}
	return nil
}
