package Authentication

import (
	"reflect"
)

//Will check if the entered struct has any uninitialized types (either nil or empty string)
func EmptyFieldsCheck(value reflect.Value) bool {
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Interface() == nil || value.Field(i).Interface() == "" {
			return false
		}
	}
	return true
}
