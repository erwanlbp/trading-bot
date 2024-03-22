package util

import (
	"encoding/json"
	"fmt"
)

func DebugPrintJson(x interface{}) {
	b, _ := json.Marshal(x)
	fmt.Println(string(b))
}

func ToJSON(x interface{}) string {
	b, _ := json.Marshal(x)
	return string(b)
}
