
import { providers } from 'ethers'
import axios from 'axios'

declare global {
  interface Window {
    ethereum: providers.ExternalProvider
  }
}

interface Config {
  getChallenge: (walletAddress: string) => Promise<string>
  login: (walletAddress: string, signedChallenge: string) => Promise<void>
  logout: () => Promise<void>
}

export class MetamaskNotInstalledError extends Error {
  constructor () {
    super('Metamask is not installed')
  }
}

async function getChallenge (walletAddress: string): Promise<string> {
  const { challenge } = (await axios.post('/auth/challenge', { walletAddress })).data
  return challenge
}

async function login(challenge: string, signature: string): Promise<void> {
  await axios.post('/auth/login', { challenge, signature })
}

async function logout (): Promise<void> {
  await axios.delete('/auth/logout')
}

export class Gatekeeper {
  provider: providers.Web3Provider
  connected: boolean
  config: Config

  constructor (config: Config) {
    if (window.ethereum === undefined) throw new MetamaskNotInstalledError()
    this.provider = new providers.Web3Provider(window.ethereum)
    this.connected = false

    const defaultConfig: Config = {
      getChallenge,
      login,
      logout
    }

    this.config = { ...defaultConfig, ...config }
  }

  async init (): Promise<void> {
    this.connected = (await this.provider.listAccounts()).length !== 0
  }

  async connectWallet (): Promise<void> {
    await this.provider.send('eth_requestAccounts', [])
    this.connected = true
  }

  get signer (): providers.JsonRpcSigner {
    return this.provider.getSigner()
  }

  async getWalletAddress (): Promise<string> {
    return await this.signer.getAddress()
  }

  async login (): Promise<void> {
    const walletAddress = await this.getWalletAddress()
    const challenge = await this.config.getChallenge(walletAddress)
    const signature = await this.signer.signMessage(challenge)
    await this.config.login(challenge, signature)
  }

  async logout (): Promise<void> {
    await this.config.logout()
  }
}

export default Gatekeeper
