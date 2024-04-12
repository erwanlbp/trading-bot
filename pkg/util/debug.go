package util

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

func DebugPrintJson(x interface{}) {
	b, _ := json.Marshal(x)
	fmt.Println(string(b))
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
