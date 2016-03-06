package dig

import "testing"

func TestAddress(t *testing.T) {
	a := Address{
		HouseNumber: "72",
		Street:      "N 4th Street",
		City:        "Brooklyn",
		Region:      "NY",
		Country:     "US",
		Postcode:    "11249"}
	b := QueryAddress("house=72&street=N 4th Street&city=Brooklyn&region=NY&country=US&postcode=11249")
	c := StringAddress("72,N 4th Street,Brooklyn,NY,US,11249")
	addrs := []Address{a, b, c}
	for _, addr := range addrs {
		if !addr.Equals(a) {
			t.Errorf("%s", addr)
		}
	}
}
