package core

import (
	"os"

	"github.com/nalanj/ladle/config"
)

// EnsureRuntimeDir creates the runtime directory, .ladle
// if it's not present
func EnsureRuntimeDir(conf *config.Config) error {
	_, statErr := os.Stat(conf.RuntimeDir())
	if os.IsNotExist(statErr) {
		return os.Mkdir(conf.RuntimeDir(), os.ModePerm)
	} else {
		return statErr
	}
}
