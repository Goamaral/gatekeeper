
import { providers } from 'ethers'

declare global {
  interface Window {
    ethereum: providers.ExternalProvider
  }
}

interface Config {
  issueChallenge: typeof issueChallenge
  register: typeof register
  login: typeof login
  logout: typeof logout
}

export class MetamaskNotInstalledError extends Error {
  constructor() {
    super('Metamask is not installed')
  }
}

export class HttpError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

export class Gatekeeper {
  provider: providers.Web3Provider
  connected: boolean
  config: Config

  constructor(config: Config) {
    if (window.ethereum === undefined) throw new MetamaskNotInstalledError()
    this.provider = new providers.Web3Provider(window.ethereum)
    this.connected = false

    const defaultConfig: Config = {
      issueChallenge,
      register,
      login,
      logout
    }

    this.config = { ...defaultConfig, ...config }
  }

  async init(): Promise<void> {
    this.connected = (await this.provider.listAccounts()).length !== 0
  }

  async connectWallet(): Promise<void> {
    await this.provider.send('eth_requestAccounts', [])
    this.connected = true
  }

  get signer(): providers.JsonRpcSigner {
    return this.provider.getSigner()
  }

  async getWalletAddress(): Promise<string> {
    return await this.signer.getAddress()
  }

  async register(email: string): Promise<void> {
    const walletAddress = await this.getWalletAddress()
    const challenge = await this.config.issueChallenge(walletAddress)
    const signature = await this.signer.signMessage(challenge)
    await this.config.register(walletAddress, challenge, signature, email)
  }

  async login(): Promise<void> {
    const walletAddress = await this.getWalletAddress()
    const challenge = await this.config.issueChallenge(walletAddress)
    const signature = await this.signer.signMessage(challenge)
    await this.config.login(walletAddress, challenge, signature)
  }

  async logout(): Promise<void> {
    await this.config.logout()
  }
}

export default Gatekeeper

/* PRIVATE */
async function sendPost<T>(url: string, body: any) {
  const res = await fetch(url, {
    method: 'POST',
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  if (res.status >= 400) throw new HttpError(res.status, (await res.json()).error)
  return (await res.json()) as T
}

async function sendDelete(url: string,) {
  const res = await fetch(url, { method: 'DELETE' })
  if (res.status >= 400) throw new HttpError(res.status, (await res.json()).error)
}

async function issueChallenge(walletAddress: string) {
  return (await sendPost<{ challenge: string }>('/auth/challenge', { walletAddress })).challenge
}

async function register(walletAddress: string, challenge: string, signature: string, email: string) {
  await sendPost('/auth/register', { walletAddress, email, challenge, signature })
}

async function login(walletAddress: string, challenge: string, signature: string) {
  await sendPost('/auth/login', { walletAddress, challenge, signature })
}

async function logout() {
  await sendDelete('/auth/logout')
}