const { SignJWT, jwtVerify } = require('jose')
const Datastore = require('nedb')
const Web3SSO = require('web3-sso')

const config = require('../config')

const db = new Datastore()

class AuthService extends Web3SSO {
  // Generate jwt
  async generateJwt(payload) {
    return await new SignJWT(payload)
      .setExpirationTime(config.jwt.expirationTime)
      .setProtectedHeader({ alg: 'ES256' })
      .sign(await config.jwt.privateKey)
  }

  // Validate jwt
  async validateJwt(jwt) {
    try {
      const { payload } = await jwtVerify(jwt, await config.jwt.publicKey)
      const current = Math.floor(new Date().getTime() / 1000)
      return { payload, valid: current < payload.exp }
    } catch (err) {
      return { payload: null, valid: false }
    }
  }
}

const MINUTE = 60 * 1000

const store = {
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

module.exports = new AuthService({ store })