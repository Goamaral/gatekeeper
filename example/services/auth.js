const { SignJWT, jwtVerify } = require('jose')
const Datastore = require('nedb')
const web3SSO = require('web3-sso')

const config = require('../config')

const db = new Datastore()

class AuthService extends web3SSO.Service {
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

module.exports = new AuthService({
  async set(walletAddress, challenge) {
    return new Promise((resolve, reject) => {
      db.update({ walletAddress }, { walletAddress, challenge }, { upsert: true }, err => {
        err ? reject(err) : resolve()
      })
    })
  },

  async get(walletAddress) {
    return new Promise((resolve, reject) => {
      db.findOne({ walletAddress }, (err, entry) => {
        if (err) {
          reject(err)
        } else if (!entry) {
          reject(new Error(`No challenge for wallet address ${walletAddress} was found`))
        } else {
          resolve(entry.challenge)
        }
      })
    })
  }
})