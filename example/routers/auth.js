const { generateRouter } = require('web3-sso/express')

const config = require('../config')
const authService = require('../services/auth')
const authMiddleware = require('../middlewares/auth')

const router = generateRouter(authService)

router.post('/login', async (req, res) => {
  try {
    const valid = await authService.validateSignedChallenge(req.body.wallet_address, req.body.signed_challenge)
    if (!valid) {
      throw new Error('Failed to validate signed challenge')
    }

    const jwt = await authService.generateJwt({ user: { walletAddress: req.body.wallet_address } })

    res.cookie('jwt', jwt, { httpOnly: true, signed: true, secure: config.env.isProduction, secret: config.cookie.secret })
    res.sendStatus(204)
  } catch (err) {
    res.status(401).json({ error: err.message })
  }
})

router.get('/user', authMiddleware, (req, res) => res.json({ user: req.user }))

module.exports = router