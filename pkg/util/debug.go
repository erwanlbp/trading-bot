package util

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

func DebugPrintJson(x ...interface{}) {
	for _, e := range x {
		b, _ := json.Marshal(e)
		fmt.Println(string(b))
	}
}

func DebugPrintYaml(x interface{}) {
	b, _ := yaml.Marshal(x)
	fmt.Println(string(b))
}

func ToJSON(x interface{}) string {
	b, _ := json.Marshal(x)
	return string(b)
}

func ToYAML(x interface{}) string {
	b, _ := yaml.Marshal(x)
	return string(b)
}
