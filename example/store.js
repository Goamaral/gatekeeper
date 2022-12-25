const MINUTE = 60 * 1000
const db = new Map()

module.exports = {
  async set (walletAddress, challenge) {
    db.set(walletAddress, { walletAddress, challenge, expirestAt: Date.now() + MINUTE })
    return Promise.resolve()
  },

  async get (walletAddress) {
    const entry = db.get(walletAddress)
    if (!entry) {
      return Promise.reject(new Error(`No challenge for wallet address ${walletAddress} was found`))
    }
    
    const { challenge, expirestAt } = entry
    // Check if challenge has expired
    if (expirestAt <= Date.now()) {
      await this.delete(walletAddress)
      return Promise.reject(new Error(`Challenge for wallet address ${walletAddress} has expired`))
    }

    return Promise.resolve(challenge)
  },

  async delete (walletAddress) {
    db.delete(walletAddress)
    return  Promise.resolve()
  }
}
