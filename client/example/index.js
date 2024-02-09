import http from 'node:http'
import fs from 'node:fs'
import crypto from 'node:crypto'

const PORT = 4000

const OK = 200
const NO_CONTENT = 204
const UNAUTHORIZED = 401
const NOT_FOUND = 404
const METHOD_NOT_ALLOWED = 405
const INTERNAL_SERVER_ERROR = 500

const sessions = []

class HttpError extends Error {
  /**
   * @param {number} status 
   * @param {string} message 
  */
  constructor(status, message) {
    super(message)
    this.status = status
  }
}

/** 
 * @param {string} walletAddress
 * @returns {Promise<string>}
 */
async function issueChallenge(walletAddress) {
  const res = await fetch('http://localhost:3000/v1/challenges/issue', {
    method: 'POST',
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ walletAddress }),
  })
  if (res.status >= 400) {
    const { error } = await res.json()
    console.error(`Error (issueChallenge):`, error)
    throw new HttpError(res.status, res.statusText)
  }

  const { challenge } = await res.json()
  return challenge
}

/**
 * @param {string} challenge 
 * @param {string} signature 
 * @returns {Promise<boolean>}
 */
async function verifyChallenge(challenge, signature) {
  const res = await fetch('http://localhost:3000/v1/challenges/verify', {
    method: 'POST',
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ challenge, signature }),
  })
  if (res.status >= 500) {
    const { error } = await res.json()
    console.error(`Error (verifyChallenge):`, error)
    throw new HttpError(res.status, res.statusText)
  }

  if (res.status >= 400) {
    const { error } = await res.json()
    console.warn(`Warn (verifyChallenge):`, error)
    return false
  }
  return true
}

const server = http.createServer(async function (req, res) {
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
      '/auth/challenge': async () => sendJson({ challenge: await issueChallenge(req.body.walletAddress) }),
      '/auth/login': async () => {
        const session = getSession()
        if (authenticated(() => session)) return sendStatus(NO_CONTENT)

        const valid = await verifyChallenge(req.body.challenge, req.body.signature)
        if (!valid) return sendJson({ error: 'failed to validate signed challenge' }, UNAUTHORIZED)

        const newSession = crypto.randomBytes(20).toString('hex')
        sessions.push(newSession)
        res.setHeader('Set-Cookie', `session=${newSession};`)
        sendStatus(NO_CONTENT)
      }
    },
    DELETE: {
      '/auth/logout': () => {
        const session = getSession()
        if (authenticated(() => session)) {
          const index = sessions.findIndex(s => s === session)
          if (index !== -1) sessions.splice(index, 1)
        }
        res.setHeader('Set-Cookie', 'session=')
        sendStatus(NO_CONTENT)
      }
    }
  }

  if (!Object.keys(router).includes(req.method)) return sendJson({ error: 'Method Not Allowed' }, METHOD_NOT_ALLOWED)

  const handler = router[req.method][req.url]
  if (handler === undefined) return sendJson({ error: 'Not Found' }, NOT_FOUND)

  try {
    await handler(req, res)
  } catch (err) {
    if (err instanceof HttpError) {
      sendJson({ error: err.message }, err.status)
    } else {
      console.error(err)
      sendStatus(INTERNAL_SERVER_ERROR)
    }
  }

  /* HELPERS */
  /**
   * @returns {Promise<string>}
   * */
  function readBody() {
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

  function getSession() {
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
  function sendFile(path, contentType = 'text/plain') {
    res.setHeader('Content-Type', contentType)
    const stream = fs.createReadStream(path)
    stream.pipe(res)
  }

  /**
   * @param {any} object
   * @param {number} status
  */
  function sendJson(object, status = OK) {
    res.statusCode = status
    res.setHeader('Content-Type', 'application/json')
    res.write(JSON.stringify(object))
    res.end()
  }

  function sendStatus(status) {
    res.statusCode = status
    res.end()
  }

  function authenticated(getSessionFn = getSession) {
    const session = getSessionFn()
    return !!session && sessions.includes(session)
  }
})

server.listen(PORT, () => console.log(`Listening on port ${PORT}`))
