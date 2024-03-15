package util

import (
	"encoding/json"
	"fmt"
)

func DebugPrintJson(x interface{}) {
	b, _ := json.Marshal(x)
	fmt.Println(string(b))
}
