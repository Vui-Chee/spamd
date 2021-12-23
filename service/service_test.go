package service

import (
	"testing"
)

func TestListenReturnsErrOnInvalidPort(t *testing.T) {
	invalidPorts := []int{
		-1,
		1,
		1023,
	}

	for _, port := range invalidPorts {
		_, err := Listen(port)
		if err == nil {
			t.Errorf("Should return error if port == %d, Got: error == nil.\n", port)
		}
	}
}
