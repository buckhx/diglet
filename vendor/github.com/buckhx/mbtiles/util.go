package mbtiles

import (
	"os"
)

func isPathAvailable(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return false
		}
	} else {
		// Path already exists
		return false
	}
	return true
}

func flipY(y, z int) int {
	return (1 << uint(z)) - y - 1
}
