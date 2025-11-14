package main

import (
	"gno.land/r/aib/ibc/core"
)

func main() {
	var (
		clientID                 = "" // XXX update
		counterpartyMerklePrefix = [][]byte{
			[]byte("iavlStoreKey"), // XXX update
			[]byte("storeKey"),     // XXX update
		}
		counterpartyClientID = "" // XXX update
	)
	core.RegisterCounterparty(cross, clientID, counterpartyMerklePrefix, counterpartyClientID)
}

// Error:
// invalid counterparty client id
