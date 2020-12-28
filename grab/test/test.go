package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"grab"
	modules2 "grab/modules"
	"log"
	"net"
	"reflect"
)

func Stringify(v interface{}) string {
	outputBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(outputBuffer)
	encoder.SetEscapeHTML(false)
	//encoder.SetIndent("", "  ")
	encoder.Encode(v)
	return outputBuffer.String()
}

type TestFlag struct {
	Name     string `default:"80"`
	Age      string `default:"8000"`
	BaseFlag struct {
		BaseName string `default:"9999"`
	}
}

// 遍历struct并且自动进行赋值
func FillDefaultValue(flag interface{}) error {
	flagType := reflect.TypeOf(flag)
	flagValue := reflect.ValueOf(flag)
	if flagType.Kind() == reflect.Ptr {
		flagType = flagType.Elem()
		flagValue = flagValue.Elem()
	} else {
		return errors.New("flag must be ptr to struct")
	}
	// 遍历结构体
	for i := 0; i < flagType.NumField(); i++ {
		typeField := flagType.Field(i)
		valueField := flagValue.Field(i)

		structType := typeField.Type
		if structType.Kind() == reflect.Struct {
			structValue := valueField.Interface()
			if err := FillDefaultValue(structValue); err != nil {
				return err
			}
			fmt.Println(structValue)
		} else {
			defaultValue := typeField.Tag.Get("default")
			if !reflect.ValueOf(defaultValue).IsZero() {
				defaultValueType := reflect.TypeOf(defaultValue)
				if structType == defaultValueType {
					valueField.Set(reflect.ValueOf(defaultValue))
				} else {
					if defaultValueType.ConvertibleTo(structType) {
						valueField.Set(reflect.ValueOf(defaultValue).Convert(structType))
					} else {
						return errors.New(typeField.Name + " type mismatch")
					}
				}
			}
		}

	}
	return nil
}

func main() {
	modules := modules2.NewModuleSetWithDefaults()
	module := modules["tls"]
	scanner := module.NewScanner()
	flags := module.NewFlags()
	scanner.Init(flags.(grab.ScanFlags))
	target := grab.ScanTarget{
		IP: net.IP{127, 0, 0, 1},
	}
	status, res, err := scanner.Scan(target)
	str := Stringify(res)
	log.Println(str)
	if err != nil {
		log.Println(status, string(str))
	} else {
		log.Println(err)
	}
}
