const crypto = require('crypto')

const BaseStore = require('./base_store')

module.exports = class {
  constructor(store = new BaseStore()) {
    // TODO: Verify store is an object
    this.store = store
  }

  // Issue challenge
  async issueChallenge(walletAddress) {
    const challenge = this._generateChallengeToken()
    await this.store.set(walletAddress, challenge)
    return challenge
  }

  // Validate signed challenge
  async validateSignedChallenge(walletAddress, signedChallenge) {
    const persistedChallenge = await this.store.get(walletAddress)
    const challenge = signedChallenge // TODO: Decrypt signed challenge
    return persistedChallenge === challenge
  }

  // Generate challenge token
  _generateChallengeToken() {
    return crypto.createHash('sha256')
      .update((crypto.randomInt(2 << 48) * new Date()).toString(10))
      .digest('hex')
  }
}