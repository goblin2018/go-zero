package mgo

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	// VarStringSliceType describes a golang data structure name for mongo.
	VarStringName string
)

// Action provides the entry for goctl mongo code generation.
func Action(_ *cobra.Command, _ []string) error {
	name := VarStringName
	if len(name) == 0 {
		return errors.New("missing name")
	}

	return GenMongo(name)
}
