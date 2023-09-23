package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSegment(t *testing.T) {
	segment, err := NewSegment("danya")
	assert.Nil(t, err)

	fmt.Printf("%+v\n", segment)
}
