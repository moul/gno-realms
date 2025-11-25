package main

import (
	"gno.land/r/aib/ibc/core"
)

// RegisterCounterparty w/o existing client
func main() {
	core.RegisterCounterparty(cross, "unknown-client", [][]byte{[]byte("iavlStoreKey"), []byte("prefix2")}, "07-tendermint-2")
}

// Error:
// client unknown-client not found
