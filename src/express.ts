import { Web3SSOBackend } from './backend.js'

export const addRoutes = (router, service = new Web3SSOBackend()): void => {
  // Using a POST to keep client wallet_address private
  router.post('/challenge', async (req, res) => {
    res.json({ challenge: await service.issueChallenge(req.body.walletAddress) })
  })
}
