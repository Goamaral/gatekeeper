import cookieParser from 'cookie-parser'
import express from 'express'
import { readFileSync } from 'fs'
import { importPKCS8, importSPKI } from 'jose'

import config from './config.js'
import router from './router.js'

function main () {
  const app = express()
  const port = 3000

  app.use(express.json())
  app.use(cookieParser(config.cookie.secret))
  app.use(express.static('public'))

  app.use('/auth', router)

  app.listen(port, () => console.log(`Example app listening at http://localhost:${port}`))
}

async function init () {
  config.jwt = {
    ...config.jwt,
    privateKey: await importPKCS8(readFileSync('./secrets/jwt.priv').toString(), 'ES256'),
    publicKey: await importSPKI(readFileSync('./secrets/jwt.pub').toString(), 'ES256')
  }

  main()
}

init()
