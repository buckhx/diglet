package dig

import "testing"

func TestMphoneIndex(t *testing.T) {
	addrs := []string{
		"42nd Street",
		"N 4th Street",
		"n. forth Street",
	}
	for _, addr := range addrs {
		mphones := mphones(addr)
		for mphone := range mphones {
			t.Errorf(mphone)
		}
	}
}

func TestExpand(t *testing.T) {
	addrs := []string{
		"72 n forth st",
		"72 n. forth st",
		"72 n. 4th st",
		"n. 4th st",
		"north 4th st",
	}
	for _, addr := range addrs {
		exaddr := expand(clean(addr))
		t.Errorf("%s -> %s ", addr, exaddr)
	}
}
