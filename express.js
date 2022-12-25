const Web3SSO = require('.')

module.exports.addRoutes = (router, service = new Web3SSO()) => {
  // Using a POST to keep client wallet_address private
  router.post('/challenge', async (req, res) => {
    res.json({ challenge: await service.issueChallenge(req.body.wallet_address) })
  })
}
