package util

import "testing"

func TestGetLookup(t *testing.T) {
	l := GetLookup()

	var i interface{} = l
	_, ok := i.(ILookup)

	if !ok {
		t.Error("GetLookup doesn't return an implementation of ILookup")
	}
}

func TestLookup_GetIP(t *testing.T) {
	l := GetLookup()

	ip, err := l.GetIP("google.com")

	if ip == "" || err != nil {
		t.Error("GetIP didn't return an IP address for Google")
	}

	ip, err = l.GetIP("ABC FTW")
	if err == nil {
		t.Error("Looking up IP for ABC FTW should have thrown an error...")
	}
}
