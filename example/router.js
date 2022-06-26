const { addRoutes } = require('web3-sso/express')
const { Router } = require('express')

const config = require('./config')
const service = require('./service')
const middleware = require('./middleware')

const router = Router()
addRoutes(router, service)

router.post('/login', async (req, res) => {
  try {
    const valid = await service.validateSignedChallenge(req.body.wallet_address, req.body.signed_challenge)
    if (!valid) {
      throw new Error('Failed to validate signed challenge')
    }

    const jwt = await service.generateJwt({ user: { walletAddress: req.body.wallet_address } })

    res.cookie('jwt', jwt, { httpOnly: true, signed: true, secure: config.env.isProduction, secret: config.cookie.secret })
    res.sendStatus(204)
  } catch (err) {
    res.status(401).json({ error: err.message })
  }
})

router.get('/user', middleware, (req, res) => res.json({ user: req.user }))

module.exports = router
