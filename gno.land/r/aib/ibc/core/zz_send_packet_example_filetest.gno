// PKGPATH: gno.lang/r/aib/main
package main

import (
	"chain/banker"

	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

func main() {
	var (
		packet = types.MsgSendPacket{
			SourceClient:     "",        // TODO update
			TimeoutTimestamp: uint64(0), // TODO update
			Payloads: []types.Payload{
				{
					SourcePort:      "",           // TODO update
					DestinationPort: "",           // TODO update
					Encoding:        "",           // TODO update
					Value:           []byte("{}"), // TODO update
					Version:         "",           // TODO update
				},
			},
		}
		banker = banker.NewBanker(banker.BankerTypeOriginSend)
	)

	sequence := core.SendPacket(cross, banker, packet)

	println(sequence)
}

// Error:
// validate client identifier: identifier cannot be blank
