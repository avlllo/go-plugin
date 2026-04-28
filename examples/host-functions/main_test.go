package main

import (
	"testing"

	"github.com/runtime-radar/go-plugin/tests"
	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	got := tests.TestStdout(t, main)
	want := `Hello, go-plugin from`
	assert.Contains(t, got, want)
}
