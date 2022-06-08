const { SignJWT, jwtVerify } = require('jose')
const Web3SSO = require('web3-sso')

const config = require('./config')
const store = require('./store')

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

module.exports = new AuthService({ store })