package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
)

// buildMsgSendPacket creates a MsgSendPacket for an IBC v2 token transfer.
func buildMsgSendPacket(sourceClient, sender, receiver, denom, amount string, timeoutTimestamp int64) *channeltypesv2.MsgSendPacket {
	packetData := transfertypes.NewFungibleTokenPacketData(denom, amount, sender, receiver, "")
	bz, err := proto.Marshal(&packetData)
	if err != nil {
		panic(fmt.Sprintf("marshal FungibleTokenPacketData: %v", err))
	}

	payload := channeltypesv2.NewPayload(
		transfertypes.PortID, transfertypes.PortID,
		transfertypes.V1, transfertypes.EncodingProtobuf, bz,
	)
	return channeltypesv2.NewMsgSendPacket(
		sourceClient, uint64(timeoutTimestamp), sender, payload,
	)
}

// buildUnsignedTx wraps messages into an unsigned Cosmos SDK Tx and returns its JSON representation.
func buildUnsignedTx(msgs []proto.Message, registerInterfaces ...func(codectypes.InterfaceRegistry)) string {
	ir := codectypes.NewInterfaceRegistry()
	for _, register := range registerInterfaces {
		register(ir)
	}
	cdc := codec.NewProtoCodec(ir)

	anyMsgs := make([]*codectypes.Any, len(msgs))
	for i, msg := range msgs {
		anyMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(fmt.Sprintf("pack message: %v", err))
		}
		anyMsgs[i] = anyMsg
	}

	tx := txtypes.Tx{
		Body: &txtypes.TxBody{
			Messages: anyMsgs,
		},
		AuthInfo: &txtypes.AuthInfo{
			Fee: &txtypes.Fee{
				Amount:   sdk.NewCoins(sdk.NewInt64Coin("uphoton", 10000)),
				GasLimit: 300000,
			},
		},
		Signatures: [][]byte{},
	}

	txJSON, err := cdc.MarshalJSON(&tx)
	if err != nil {
		panic(fmt.Sprintf("marshal tx: %v", err))
	}
	return string(txJSON)
}

// recoverGnoKey recovers a key in gnokey inside the gno container from a mnemonic.
// If the key already exists, it is a no-op.
func recoverGnoKey(containerID, keyName, mnemonic string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// gnokey add --recover reads: mnemonic first, then passphrase (empty = unencrypted)
	stdin := fmt.Sprintf("%s\n\n", mnemonic)
	_, stderr, err := dockerExecStdin(ctx, containerID, stdin,
		"gnokey", "add", keyName, "--recover", "--insecure-password-stdin", "--force")
	if err != nil {
		return fmt.Errorf("gnokey add --recover: %w: %s", err, stderr)
	}
	return nil
}

// gnoKeyAddress returns the address associated with a gnokey key name.
func gnoKeyAddress(containerID, keyName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stdout, stderr, err := dockerExec(ctx, containerID, "gnokey", "list")
	if err != nil {
		return "", fmt.Errorf("gnokey list: %w: %s", err, stderr)
	}
	// Output format: "0. keyname (local) - addr: g1... pub: gpub1..."
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, keyName) {
			idx := strings.Index(line, "addr: ")
			if idx >= 0 {
				rest := line[idx+len("addr: "):]
				addr := strings.Fields(rest)[0]
				return addr, nil
			}
		}
	}
	return "", fmt.Errorf("key %s not found in gnokey list output: %s", keyName, stdout)
}

// signAndBroadcastGnoCall executes a gnokey maketx call inside the gno container.
func signAndBroadcastGnoCall(containerID, chainID, keyName, pkgPath, funcName, sendCoins string, args ...string) error {
	cmdArgs := []string{
		"gnokey", "maketx", "call",
		"-pkgpath", pkgPath,
		"-func", funcName,
		"-gas-fee", "1000000ugnot",
		"-gas-wanted", "90000000",
		"-broadcast",
		"-chainid", chainID,
		"-remote", "localhost:26657",
		"-insecure-password-stdin",
	}
	if sendCoins != "" {
		cmdArgs = append(cmdArgs, "-send", sendCoins)
	}
	for _, arg := range args {
		cmdArgs = append(cmdArgs, "-args", arg)
	}
	cmdArgs = append(cmdArgs, keyName)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	stdout, stderr, err := dockerExecStdin(ctx, containerID, "\n", cmdArgs...)
	if err != nil {
		return fmt.Errorf("gnokey maketx call: %w: stdout=%s stderr=%s", err, stdout, stderr)
	}
	return nil
}

// signAndBroadcastAtomOneTx signs and broadcasts messages on the AtomOne chain.
// It returns the tx hash.
func signAndBroadcastAtomOneTx(containerID, chainID string, msgs ...proto.Message) (string, error) {
	unsignedTx := buildUnsignedTx(msgs, channeltypesv2.RegisterInterfaces)

	ctx := context.Background()

	// Write unsigned tx to container
	_, stderr, err := dockerExecStdin(ctx, containerID, unsignedTx,
		"bash", "-c", "cat > /tmp/unsigned_tx.json")
	if err != nil {
		return "", fmt.Errorf("write unsigned tx: %w: %s", err, stderr)
	}

	// Sign
	signCtx, signCancel := context.WithTimeout(ctx, 30*time.Second)
	defer signCancel()
	_, stderr, err = dockerExec(signCtx, containerID,
		"atomoned", "tx", "sign", "/tmp/unsigned_tx.json",
		"--from", "validator",
		"--chain-id", chainID,
		"--keyring-backend", "test",
		"--home", "/root/.atomone",
		"--node", "tcp://localhost:26657",
		"--output-document", "/tmp/signed_tx.json",
	)
	if err != nil {
		return "", fmt.Errorf("sign tx: %w: %s", err, stderr)
	}

	// Broadcast
	bcastCtx, bcastCancel := context.WithTimeout(ctx, 30*time.Second)
	defer bcastCancel()
	stdout, stderr, err := dockerExec(bcastCtx, containerID,
		"atomoned", "tx", "broadcast", "/tmp/signed_tx.json",
		"--node", "tcp://localhost:26657",
		"--output", "json",
	)
	if err != nil {
		return "", fmt.Errorf("broadcast tx: %w: %s", err, stderr)
	}

	var txResult struct {
		Code   int    `json:"code"`
		TxHash string `json:"txhash"`
		RawLog string `json:"raw_log"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &txResult); err != nil {
		return "", fmt.Errorf("parse broadcast result: %w", err)
	}
	if txResult.Code != 0 {
		return "", fmt.Errorf("tx failed (code %d): %s", txResult.Code, txResult.RawLog)
	}
	return txResult.TxHash, nil
}
