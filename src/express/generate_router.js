const router = require('express').Router()

const Service = require('../service')

module.exports = (service = new Service()) => {
  // Using a POST to keep client walletAddress private
  router.post('/challenge', async (req, res) => {
    res.json({ challenge: await service.issueChallenge(req.body.wallet_address) })
  })

  return router
}