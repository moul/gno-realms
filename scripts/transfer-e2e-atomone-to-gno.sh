#!/usr/bin/env bash
# Send tokens via IBC v2 MsgSendPacket from AtomOne to Gno using the e2e localnet.
#
# Defaults: 20 ATONE (20_000_000 uatone) atone1z437…→g1z437… via 10-gno-0.
# Override via env: DENOM, AMOUNT, RECEIVER, CLIENT_ID, SIGNER, MEMO, TIMEOUT.
#
# Usage:
#   ./scripts/transfer-atomone-to-gno.sh
#   AMOUNT=5000000 DENOM=uphoton ./scripts/transfer-atomone-to-gno.sh

set -euo pipefail

CONTAINER="${CONTAINER:-e2e-atomone-1}"
CHAIN_ID="${CHAIN_ID:-atomone-e2e-1}"
CLIENT_ID="${CLIENT_ID:-10-gno-0}"
SIGNER="${SIGNER:-validator}"
SENDER="${SENDER:-atone1z437dpuh5s4p64vtq09dulg6jzxpr2hdgu88r6}"
RECEIVER="${RECEIVER:-g1z437dpuh5s4p64vtq09dulg6jzxpr2hd4q8r5x}"
#SENDER="${SENDER:-atone12j6x2cnpkvz83l6a5lhfw22703kwwpknu4n0fn}" # aib testnet account
#RECEIVER="${RECEIVER:-g12j6x2cnpkvz83l6a5lhfw22703kwwpknpfnt70}" # aib testnet account
DENOM="${DENOM:-uatone}"
AMOUNT="${AMOUNT:-20000000}"
MEMO="${MEMO:-}"
TIMEOUT="${TIMEOUT:-$(( $(date +%s) + 3600 ))}"

# Encode FungibleTokenPacketData proto: denom(1), amount(2), sender(3), receiver(4), memo(5)
PAYLOAD_B64=$(python3 - "$DENOM" "$AMOUNT" "$SENDER" "$RECEIVER" "$MEMO" <<'PY'
import base64, sys
def f(tag, s):
    b = s.encode()
    out = bytes([tag<<3 | 2])
    n = len(b)
    while n > 0x7f:
        out += bytes([n & 0x7f | 0x80]); n >>= 7
    out += bytes([n]) + b
    return out
denom, amount, sender, receiver, memo = sys.argv[1:6]
buf = f(1, denom) + f(2, amount) + f(3, sender) + f(4, receiver)
if memo:
    buf += f(5, memo)
print(base64.b64encode(buf).decode())
PY
)

UNSIGNED=$(cat <<JSON
{
  "body": {
    "messages": [{
      "@type": "/ibc.core.channel.v2.MsgSendPacket",
      "source_client": "$CLIENT_ID",
      "timeout_timestamp": "$TIMEOUT",
      "payloads": [{
        "source_port": "transfer",
        "destination_port": "transfer",
        "version": "ics20-1",
        "encoding": "application/x-protobuf",
        "value": "$PAYLOAD_B64"
      }],
      "signer": "$SENDER"
    }],
    "memo": "",
    "timeout_height": "0"
  },
  "auth_info": {
    "fee": {
      "amount": [{"denom": "uphoton", "amount": "10000"}],
      "gas_limit": "300000"
    },
    "signer_infos": []
  },
  "signatures": []
}
JSON
)

echo "==> Transferring $AMOUNT $DENOM from $SENDER -> $RECEIVER via $CLIENT_ID (timeout=$TIMEOUT)"
echo "$UNSIGNED" | docker exec -i "$CONTAINER" sh -c 'cat > /tmp/unsigned.json'
docker exec "$CONTAINER" atomoned tx sign /tmp/unsigned.json \
    --from "$SIGNER" --chain-id "$CHAIN_ID" --keyring-backend test \
    --home /root/.atomone --node tcp://localhost:26657 \
    --output-document /tmp/signed.json
docker exec "$CONTAINER" atomoned tx broadcast /tmp/signed.json \
    --node tcp://localhost:26657 --output json
