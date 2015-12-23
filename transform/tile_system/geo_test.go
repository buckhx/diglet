package tile_system

import (
	"testing"
)

var clipTests = []struct {
	val, min, max, out float64
	hasErr             bool
}{
	{0, 1, 5, 1, false},
	{9, 1, 5, 5, false},
	{3, 1, 5, 3, false},
	{7, 5, 1, 0, true},
}

func TestClip(t *testing.T) {
	for _, test := range clipTests {
		val, err := clip(test.val, test.min, test.max)
		hasErr := err != nil
		if val != test.out || hasErr != test.hasErr {
			t.Errorf("clip() %+v", test)
		}

	}
}
