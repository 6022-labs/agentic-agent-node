import {
  BrowserProvider,
  formatEther,
  parseEther,
  type Eip1193Provider,
} from "ethers"

import { getScaledTxGasOverrides } from "./tx-gas-overrides"

// Read an address' native-token balance (in ether) via the connected wallet.
export const getNativeBalance = async (
  address: string,
  walletProvider: Eip1193Provider
) => {
  const provider = new BrowserProvider(walletProvider)
  return formatEther(await provider.getBalance(address))
}

// Top the address up to minEther only when it's short; returns true if it sent, letting self-mint/self-ENS skip the donation when already funded.
export const ensureGas = async (
  address: string,
  minEther: string,
  walletProvider: Eip1193Provider
): Promise<boolean> => {
  const balance = await getNativeBalance(address, walletProvider)
  if (Number(balance) >= Number(minEther)) return false
  await sendNative(address, minEther, walletProvider)
  return true
}

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
