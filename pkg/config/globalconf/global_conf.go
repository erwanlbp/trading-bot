package globalconf

import (
	"context"
	"io"
	"os"
	"strconv"
)

// To inject functions that can act on global config/dependancies object
type GlobalConfModifier interface {
	ReloadConfigFile(context.Context) error
	ExportDBFile() (io.Reader, error)
	GetDBSize() (int64, error)
}

func IsBacktesting() bool {
	val, _ := strconv.ParseBool(os.Getenv("BACKTESTING"))
	return val
}
