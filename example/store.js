const Datastore = require('nedb')

const MINUTE = 60 * 1000
const db = new Datastore()

module.exports = {
  async set(walletAddress, challenge) {
    return new Promise((resolve, reject) => {
      const data = { walletAddress, challenge, expirestAt: Date.now() + MINUTE }
      db.update({ walletAddress }, data, { upsert: true }, err => {
        err ? reject(err) : resolve()
      })
    })
  },

  async get(walletAddress) {
    return new Promise((resolve, reject) => {
      db.findOne({ walletAddress }, async (err, entry) => {
        if (err) {
          reject(err)
        } else if (!entry) {
          reject(new Error(`No challenge for wallet address ${walletAddress} was found`))
        } else {
          const { challenge, expirestAt } = entry

          // Check if challenge has expired
          if (expirestAt <= Date.now()) {
            await this.delete(walletAddress)
            reject(new Error(`Challenge for wallet address ${walletAddress} has expired`))
          }

          resolve(challenge)
        }
      })
    })
  },

  async delete(walletAddress) {
    return new Promise((resolve, reject) => {
      db.remove({ walletAddress }, {}, err => {
        err ? reject(err) : resolve()
      })
    })
  }
}