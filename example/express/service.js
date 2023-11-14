import { SignJWT, jwtVerify } from 'jose'
import Web3SSOBackend from 'web3-sso/backend'

import config from './config.js'
import store from './store.js'

class AuthService extends Web3SSOBackend {
  // Generate jwt
  async generateJwt (payload) {
    return await new SignJWT(payload)
      .setExpirationTime(config.jwt.expirationTime)
      .setProtectedHeader({ alg: 'ES256' })
      .sign(await config.jwt.privateKey)
  }

  // Validate jwt
  async validateJwt (jwt) {
    try {
      const { payload } = await jwtVerify(jwt, await config.jwt.publicKey)
      const current = Math.floor(new Date().getTime() / 1000)
      return { payload, valid: current < payload.exp }
    } catch (err) {
      return { payload: null, valid: false }
    }
  }
}

export default new AuthService({ store })
