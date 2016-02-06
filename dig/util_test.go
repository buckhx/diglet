package dig

import "testing"

func TestMphoneIndex(t *testing.T) {
	street := "42nd Street"
	for k := range mphones(street) {
		t.Error(k)
	}
	street = "N 4th Street"
	for k := range mphones(street) {
		t.Error(k)
	}
	street = "N forth Street"
	for k := range mphones(street) {
		t.Error(k)
	}
}
