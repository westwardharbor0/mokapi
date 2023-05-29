package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefinitions_Add(t *testing.T) {
	defs := Definitions{Path: ""}
	defs.Endpoints = make(map[string]*Definition)
	def := Definition{
		Endpoint:           "/test",
		Method:             "POST",
		ResponseStatusCode: 200,
	}
	require.Nil(t, defs.Add(&def))
	require.NotNil(t, defs.Add(&def))
}
