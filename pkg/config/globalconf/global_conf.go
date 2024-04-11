package globalconf

import "context"

// To inject functions that can act on global config/dependancies object
type GlobalConfModifier interface {
	ReloadConfigFile(context.Context) error
}
