const service = require('./service')

// Validate jwt and send unauthenticated response if invalid
module.exports = async (req, res, next) => {
  const { jwt } = req.signedCookies

  // Validate jwt
  const { payload, valid } = await service.validateJwt(jwt)
  if (!valid) {
    res.status(401).json({ error: "Invalid JWT" })
  } else {
    req.user = payload.user
    next()
  }
}