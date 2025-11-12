package main

import (
	appstesting "gno.land/r/aib/ibc/apps/testing"
	"gno.land/r/aib/ibc/core"
)

// RegisterApp error when already registered
func main() {
	core.RegisterApp(cross, "app1", appstesting.NewApp(cross))
	core.RegisterApp(cross, "app1", appstesting.NewApp(cross))
}

// Error:
// portID app1 has already been registered
