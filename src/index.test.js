const { Wallet } = require('ethers')

const { default: Web3SSO } = require('./index')

const challengeToken = 'challengeToken'
const challengeMessage = 'challengeMessage'
const wallet = Wallet.createRandom()

describe('Web3SSO', () => {
  let service

  beforeEach(() => {
    service = new Web3SSO({
      store: {
        get: jest.fn(),
        set: jest.fn(),
        delete: jest.fn()
      },
      generateChallengeToken: jest.fn(() => challengeToken),
      challengeMessage
    })
  })

  describe('issueChallenge', () => {
    it('should not throw any exceptions using default options', async () => {
      service = new Web3SSO()
      await service.issueChallenge(wallet.address)
    })

    it('should return the expected challenge', async () => {
      const challenge = await service.issueChallenge(wallet.address)
      expect(challenge).toBe(`${challengeMessage}${challengeToken}`)
    })
  })

  describe('validateSignedChallenge', () => {
    const validChallenge = 'validChallenge'

    beforeEach(async () => {
      service.store.get.mockImplementation(async walletAddress => {
        if (walletAddress === wallet.address) {
          return Promise.resolve(validChallenge)
        } else {
          return Promise.resolve(null)
        }
      })
    })

    it('should return true and delete the challenge if the challenge is valid', async () => {
      const validSignedChallenge = await wallet.signMessage(validChallenge)
      const isValid = await service.validateSignedChallenge(wallet.address, validSignedChallenge)
      expect(isValid).toBe(true)
      expect(service.store.delete).toHaveBeenCalledWith(wallet.address)
    })

    it('should return false if the challenge is invalid', async () => {
      const invalidSignedChallenge = await wallet.signMessage('invalidChallenge')
      const isValid = await service.validateSignedChallenge(wallet.address, invalidSignedChallenge)
      expect(isValid).toBe(false)
    })

    it('should return false if wallet addres has no challenge', async () => {
      const anotherWallet = await Wallet.createRandom()
      const isValid = await service.validateSignedChallenge(anotherWallet.address, null)
      expect(isValid).toBe(false)
    })
  })
})
