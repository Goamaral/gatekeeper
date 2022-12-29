import Web3SSO from '.'

export const addRoutes = (router, service = new Web3SSO()): void => {
  // Using a POST to keep client wallet_address private
  router.post('/challenge', async (req, res) => {
    res.json({ challenge: await service.issueChallenge(req.body.wallet_address) })
  })
}
