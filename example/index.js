import http from 'node:http'
import fs from 'node:fs'
import crypto from 'node:crypto'

import Web3SSOBackend from 'web3-sso/backend'

const PORT = 3000

const OK = 200
const NO_CONTENT = 204
const UNAUTHORIZED = 401
const NOT_FOUND = 404
const METHOD_NOT_ALLOWED = 405

const service = new Web3SSOBackend()

const sessions = []

const server = http.createServer(async function (req, res) {
  if (req.method !== 'GET' && req.method !== 'POST') return sendText(res, METHOD_NOT_ALLOWED, 'METHOD NOT ALLOWED')
  if (req.method === 'POST') req.body = await readBody()

  const router = {
    GET: {
      '/': () => sendFile('public/index.html', 'text/html'),
      '/index.js': () => sendFile('public/index.js', 'text/javascript'),
      '/index.css': () => sendFile('public/index.css', 'text/css'),
      '/auth/user': () => {
        if (!authenticated()) return sendJson({ error: 'user not logged in' }, UNAUTHORIZED)
        sendJson({ user: { id: 2023, name: 'web3' } })
      }
    },
    POST: {
      '/auth/challenge': async () => sendJson({ challenge: await service.issueChallenge(req.body.walletAddress) }),
      '/auth/login': async () => {
        const session = getSession()
        if (authenticated(() => session)) return sendNoContent()

        const valid = await service.validateSignedChallenge(req.body.walletAddress, req.body.signedChallenge)
        if (!valid) return sendJson({ error: 'failed to validate signed challenge' }, UNAUTHORIZED)

        const newSession = crypto.randomBytes(20).toString('hex')
        sessions.push(newSession)
        res.setHeader('Set-Cookie', `session=${newSession};`)
        sendNoContent()
      },
      '/auth/logout': () => {
        const session = getSession()
        if (authenticated(() => session)) {
          const index = sessions.findIndex(s => s === session)
          if (index !== -1) sessions.splice(index, 1)
        }
        res.setHeader('Set-Cookie', 'session=')
        sendNoContent()
      }
    }
  }

  const handler = router[req.method][req.url]
  if (handler === undefined) return sendText('NOT FOUND', NOT_FOUND)

  handler(req, res)

  /* HELPERS */
  /**
   * @returns {Promise<string>}
   * */
  function readBody () {
    return new Promise((resolve, reject) => {
      try {
        let body = ''
        req.on('data', chunk => { body += chunk.toString() })
        req.on('end', () => resolve(body ? JSON.parse(body) : {}))
      } catch (error) {
        reject(error)
      }
    })
  }

  function getSession () {
    let session
    req.headers.cookie.split('; ').forEach(kv => {
      const [k, v] = kv.split('=')
      if (k === 'session') session = v
    })
    return session
  }

  /**
   * @param {string} path
   */
  function sendFile (path, contentType = 'text/plain') {
    res.setHeader('Content-Type', contentType)
    const stream = fs.createReadStream(path)
    stream.pipe(res)
  }

  /**
   * @param {string} text
   * @param {number} status
   */
  function sendText (text, status = OK) {
    res.statusCode = status
    res.write(text)
    res.end()
  }

  /**
   * @param {any} object
   * @param {number} status
  */
  function sendJson (object, status = OK) {
    res.statusCode = status
    res.setHeader('Content-Type', 'application/json')
    res.write(JSON.stringify(object))
    res.end()
  }

  function sendNoContent () {
    res.statusCode = NO_CONTENT
    res.end()
  }

  function authenticated (getSessionFn = getSession) {
    const session = getSessionFn()
    return !!session && sessions.includes(session)
  }
})

server.listen(PORT, () => console.log(`Listening on port ${PORT}`))
