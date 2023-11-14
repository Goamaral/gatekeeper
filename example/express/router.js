import { addRoutes } from 'web3-sso/express'
import { Router } from 'express'

import config from './config.js'
import service from './service.js'
import middleware from './middleware.js'

const router = Router()
addRoutes(router, service)

router.post('/login', async (req, res) => {
  try {
    const valid = await service.validateSignedChallenge(req.body.walletAddress, req.body.signedChallenge)
    if (!valid) {
      throw new Error('Failed to validate signed challenge')
    }

    const jwt = await service.generateJwt({ user: { walletAddress: req.body.walletAddress } })

    res.cookie('jwt', jwt, { httpOnly: true, signed: true, secure: config.env.isProduction, secret: config.cookie.secret })
    res.sendStatus(204)
  } catch (err) {
    res.status(401).json({ error: err.message })
  }
})

router.get('/user', middleware, (req, res) => res.json({ user: req.user }))

export default router
