# Contract fixtures

`artifacts/*.json` are trimmed hardhat artifacts (`{contractName, abi, bytecode}`)
that these integration tests deploy onto a throwaway anvil chain. They are
committed on purpose.

## What is here

`TestERC20` and `TestERC20Permit` — throwaway tokens the validator probes: a
plain one for the permit2 path and a permit one (ERC-5267 `eip712Domain`) for the
EIP-3009 path. Their bytecode is frozen on purpose — ERC-20 semantics do not
move, so there is nothing to regenerate and no source is kept.

The reject cases need no artifact: an address with no code is any unused address,
and an address holding code that is not a token comes from
`container_builders.CreateContractlessAddress`, which plants a bare STOP opcode.

## Refreshing

Not expected. If a token ever has to change, compile it with any solc 0.8.28 that
resolves OpenZeppelin 5.x and keep only the three fields:

```sh
python3 -c "
import json, sys
a = json.load(open(sys.argv[1]))
json.dump({'contractName': a['contractName'], 'abi': a['abi'], 'bytecode': a['bytecode']}, open(sys.argv[2], 'w'), indent=2)
" <hardhat>/artifacts/contracts/<Name>.sol/<Name>.json artifacts/<Name>.json
```
