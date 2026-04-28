package main

import (
	"testing"

	"github.com/runtime-radar/go-plugin/tests"
	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	got := tests.TestStdout(t, main)
	want := `File loading...
Hello WASI!`
	assert.Equal(t, want, got)
}
