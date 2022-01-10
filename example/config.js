const dotenv = require('dotenv')

dotenv.config({ path: './secrets/.env' })

const config = {
  cookie: {
    secret: process.env.COOKIE_SECRET,
  },
  env: {
    isProduction: process.env.NODE_ENV === 'production',
  },
  jwt: {
    expirationTime: '1h',
    privateKey: '<JWT_PRIVATE_KEY>',
    publicKey: '<JWT_PUBLIC_KEY>',
  },
}

module.exports = config