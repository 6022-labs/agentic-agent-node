import { BrowserProvider, parseEther, type Eip1193Provider } from "ethers"

import { getScaledTxGasOverrides } from "./tx-gas-overrides"

// Send a native-token (gas) donation from the connected browser wallet to the
// agent's node wallet, so the node can self-mint + ENS-register.
export const sendNative = async (
  to: string,
  amountEther: string,
  walletProvider: Eip1193Provider
) => {
  const provider = new BrowserProvider(walletProvider)
  const signer = await provider.getSigner()
  const tx = await signer.sendTransaction({
    to,
    value: parseEther(amountEther),
    ...(await getScaledTxGasOverrides(signer)),
  })
  await tx.wait()
  return tx
}
