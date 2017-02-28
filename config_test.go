package main

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	conf := DefaultConfig()
	if err := conf.Validate(); err != nil {
		t.Error(err)
	}
}
