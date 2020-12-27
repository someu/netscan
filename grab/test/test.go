package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Age      string
	BaseFlag struct {
		BaseName string `default:"9999"`
	}
}

func MakeDefaultFlag(flag TestFlag) interface{} {
	flagType := reflect.TypeOf(flag)
	flagValue := reflect.ValueOf(&flag).Elem()
	for i := 0; i < flagType.NumField(); i++ {
		filedType := flagType.Field(i)
		filedValue := flagValue.Field(i)
		defaultValue := filedType.Tag.Get("default")
		if defaultValue != "" {
			filedValue.SetString(defaultValue)
		}
	}
	return flag
}

func main() {
	//module := mongodb.Module{}
	//scanner := module.NewScanner()
	//flags := module.NewFlags()
	//scanner.Init(flags.(grab.ScanFlags))
	//target := grab.ScanTarget{
	//	IP: net.IP{127, 0, 0, 1},
	//}
	//status, res, err := scanner.Scan(target)
	//str := Stringify(res)
	//log.Println(str)
	//if err != nil {
	//	log.Println(status, string(str))
	//} else {
	//	log.Println(err)
	//}
	flag := MakeDefaultFlag(TestFlag{})
	fmt.Println(flag)
}
