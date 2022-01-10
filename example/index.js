const cookieParser = require('cookie-parser')
const express = require('express')
const { readFileSync } = require('fs')
const { importPKCS8, importSPKI } = require('jose')

const config = require('./config')
const routers = require('./routers')

const main = () => {
  const app = express()
  const port = 3000

  app.use(express.json())
  app.use(cookieParser(config.cookie.secret))
  app.use(express.static('public'))

  app.use('/auth', routers.auth)
  app.use('/private', routers.private)

  app.listen(port, () => console.log(`Example app listening at http://localhost:${port}`))
}

const init = async () => {
  config.jwt = {
    ...config.jwt,
    privateKey: await importPKCS8(readFileSync('./secrets/jwt.priv').toString(), 'ES256'),
    publicKey: await importSPKI(readFileSync('./secrets/jwt.pub').toString(), 'ES256')
  }

  main()
}

init()