const router = require('express').Router()

const Web3SSO = require('.')

module.exports.generateRouter = (service = new Web3SSO()) => {
  // Using a POST to keep client walletAddress private
  router.post('/challenge', async (req, res) => {
    res.json({ challenge: await service.issueChallenge(req.body.wallet_address) })
  })

  return router
}