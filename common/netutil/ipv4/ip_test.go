package ipv4

import (
	"testing"
)

func TestFind(t *testing.T) {
	pattern := "0.0.0.0"
	ipSignal, err := FindOne(pattern)
	if err != nil {
		t.Errorf("Find error: %v", err)
	}
	t.Logf("found one ip for `%s': %v", pattern, ipSignal)

	pattern = "127.*"
	ip, err := FindAll(pattern)
	if err != nil {
		t.Errorf("Find error: %v", err)
	}
	t.Logf("found ip for `%s': %v", pattern, ip)

	pattern = "127.0.0.1"
	ip, err = FindAll(pattern)
	if err != nil {
		t.Errorf("Find error: %v", err)
	}
	t.Logf("found ip for `%s': %v", pattern, ip)
}
