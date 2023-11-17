
import { providers } from 'ethers'
import axios from 'axios'

declare global {
  interface Window {
    ethereum: providers.ExternalProvider
  }
}

interface Config {
  challengeUrl: string
  loginUrl: string
}

export const MetamaskNotInstalledError = class MetamaskNotInstalledError extends Error {
  constructor () {
    super('Metamask is not installed')
  }
}

export const Web3SSOFrontend = class Web3SSOFrontend {
  provider: providers.Web3Provider
  connected: boolean
  config: Config

  constructor (config) {
    if (window.ethereum === undefined) throw new MetamaskNotInstalledError()
    this.provider = new providers.Web3Provider(window.ethereum)
    this.connected = false

    const defaultConfig = {
      challengeUrl: '/auth/challenge',
      loginUrl: '/auth/login'
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
    const { challenge } = (await axios.post(this.config.challengeUrl, { walletAddress })).data
    const signedChallenge = await this.signer.signMessage(challenge)
    await axios.post(this.config.loginUrl, { walletAddress, signedChallenge })
  }
}

export default Web3SSOFrontend
