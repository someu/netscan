package grab

import (
	"fmt"
	"reflect"
)

func MakeDefaultFlag(flag interface{}) interface{} {
	flagType := reflect.TypeOf(flag)
	flagValue := reflect.ValueOf(&flag).Elem()
	for i := 0; i < flagType.NumField(); i++ {
		filedType := flagType.Field(i)
		filedValue := flagValue.Field(i)
		defaultValue := filedType.Tag.Get("default")
		fmt.Println(filedValue.CanAddr(), filedValue.CanSet())
		if defaultValue != "" {
			filedValue.SetString(defaultValue)
		}
	}
	return flag
}
