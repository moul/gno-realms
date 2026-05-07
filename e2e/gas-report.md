# E2E Gas Cost Report

| Operation | Old gas model | New gas model | Ratio |
|---|---|---|---|
| CreateClient | 36,634,635 | 61,264,946 | 1.67x |
| RegisterCounterparty | 23,896,178 | 41,196,027 | 1.72x |
| UpdateClient | ~45,000,000 | ~69,500,000 | 1.54x |
| RecvPacket | ~60,000,000 | ~98,000,000 | 1.63x |
| RecvPacket (ack) | ~50,000,000 | ~83,000,000 | 1.66x |
| call:Transfer (IBC) | ~37,000,000 | ~52,000,000 | 1.41x |
| call:Mint (GRC20) | 4,901,238 | 4,915,434 | 1.00x |
| call:Approve (GRC20) | 3,792,999 | 4,853,243 | 1.28x |
