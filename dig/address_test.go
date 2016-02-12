package dig

import "testing"

func TestAddress(t *testing.T) {
	a := Address{
		HouseNumber: "72",
		Street:      "N 4th Street",
		City:        "Brooklyn",
		Region:      "New York",
		Country:     "United States",
		Postcode:    "11249"}
	addrs := []Address{a}
	for _, addr := range addrs {
		t.Errorf("%s", addr)
	}
}
