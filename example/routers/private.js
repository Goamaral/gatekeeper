const router = require('express').Router()

const authMiddleware = require('../middlewares/auth')

router.use(authMiddleware) 
router.get('/', (req, res) => res.json({ secret: req.user.walletAddress }))

module.exports = router