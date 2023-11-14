import express from 'express'
import request from 'supertest'
import { Wallet } from 'ethers'

import { addRoutes } from './express'

const wallet = Wallet.createRandom()
let app

describe('Express', () => {
  beforeEach(() => {
    app = express()
    app.use(express.json())
  })

  describe('addRoutes', () => {
    beforeEach(() => addRoutes(app))

    it('POST /challenge should not throw any exceptions', async () => {
      const res = await request(app)
        .post('/challenge')
        .send({ wallet_address: wallet.address })

      expect(res.body).toEqual(
        expect.objectContaining({
          challenge: expect.any(String)
        })
      )
    })
  })
})
