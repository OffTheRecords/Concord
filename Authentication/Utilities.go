package Authentication

import (
	"log"
	"reflect"
	"regexp"
)

//Will check if the entered struct has any uninitialized types (either nil or empty string)
func EmptyFieldsCheck(value reflect.Value) bool {
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Interface() == nil || value.Field(i).Interface() == "" {
			return false
		}
		if value.Field(i).Interface() == "email" {

		}
	}
	return true
}

//Check if field is valid, true equals match
func FieldRegexCheck(value reflect.Value, fieldRegex string) bool {
	argMatched, err := regexp.MatchString(fieldRegex, value.Field())
	if err != nil {
		log.Fatal("Bad regex string in FieldRegexCheck")
		return false
	}

	return argMatched
}
