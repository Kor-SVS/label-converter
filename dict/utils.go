package dict

import (
	"os"
	"path/filepath"
)

func getExecutablePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(ex)
}
