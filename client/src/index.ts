import crypto from 'crypto'
import ethers from 'ethers'

function defaultGenerateChallengeToken (): string {
  return crypto.randomBytes(10).toString('hex')
}

export interface Store {
  get: (key: string) => Promise<string | undefined>
  set: (key: string, value: string) => Promise<void>
  delete: (key: string) => Promise<void>
}

class MapStore implements Store {
  map: Map<string, string>

  constructor () {
    this.map = new Map()
  }

  async get (key: any): Promise<string | undefined> {
    return await Promise.resolve(this.map.get(key))
  }

  async set (key: any, value: any): Promise<void> {
    this.map.set(key, value)
    return await Promise.resolve()
  }

  async delete (key: string): Promise<void> {
    this.map.delete(key)
    return await Promise.resolve()
  }
}

interface ConstructorOptions {
  store: Store | undefined
  generateChallengeToken: (() => string) | undefined
  challengeMessage: string | undefined
}

export const Gatekeeper = class Gatekeeper {
  store: Store
  generateChallengeToken: () => string
  challengeMessage: string

  constructor (opts?: ConstructorOptions) {
    opts = { store: undefined, generateChallengeToken: undefined, challengeMessage: undefined, ...opts }
    this.store = opts.store ?? new MapStore()
    this.generateChallengeToken = opts.generateChallengeToken ?? defaultGenerateChallengeToken
    this.challengeMessage = opts.challengeMessage ?? 'Login request\n'
  }

  async issueChallenge (walletAddress: string): Promise<string> {
    const challenge = `${this.challengeMessage}${this.generateChallengeToken()}`
    await this.store.set(walletAddress, challenge)
    return challenge
  }

  async validateSignedChallenge (walletAddress: string, signedChallenge: string): Promise<boolean> {
    // Get challenge from store
    const challenge = await this.store.get(walletAddress)
    if (challenge === undefined) return false

    // Get signer address from signed challenge
    const signerAddress = ethers.utils.verifyMessage(challenge, signedChallenge)

    // Check if signer address matches wallet address
    if (signerAddress === walletAddress) {
      await this.store.delete(walletAddress)
      return true
    } else {
      return false
    }
  }
}

export default Gatekeeper
