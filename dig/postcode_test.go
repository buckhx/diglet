package dig_test

import (
	"github.com/buckhx/diglet/dig"
	"github.com/buckhx/diglet/util"
	"testing"
)

const pc_path = "/vagrant/postcodes/allCountries.txt"

func testReadPostcodes(t *testing.T) {
	postcodes := dig.ReadPostcodes(pc_path)
	for pc := range postcodes {
		util.Info("%+v", pc)
	}
}
