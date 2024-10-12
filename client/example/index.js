import http from 'node:http'
import fs from 'node:fs'
import crypto from 'node:crypto'
import cookie from 'cookie'

const PORT = 4000
const API_KEY = "018df6ccab907592ae2da5c3dd9a79f3AFF3MAUaKHt9DVuBBi4Jzw"

const OK = 200
const NO_CONTENT = 204
const UNAUTHORIZED = 401
const NOT_FOUND = 404
const METHOD_NOT_ALLOWED = 405
const INTERNAL_SERVER_ERROR = 500

const sessions = {}

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
    headers: { "Content-Type": "application/json", "Api-Key": API_KEY },
    body: JSON.stringify({ walletAddress }),
  })
  if (res.status >= 400) {
    const { error } = await res.json()
    console.error(`Error (issueChallenge):`, error)
    throw new HttpError(res.status, error)
  }

  const { challenge } = await res.json()
  return challenge
}

/**
 * @param {string} challenge 
 * @param {string} signature 
 * @returns {Promise<{walletAddress: string, proofToken: string}|null>}
 */
async function verifyChallenge(challenge, signature) {
  const res = await fetch('http://localhost:3000/v1/challenges/verify', {
    method: 'POST',
    headers: { "Content-Type": "application/json", "Api-Key": API_KEY },
    body: JSON.stringify({ challenge, signature }),
  })
  if (res.status >= 400) {
    const { error } = await res.json()
    console.error(`Error (verifyChallenge):`, error)
    throw new HttpError(res.status, error)
  }

  const { proofToken, walletAddress } = await res.json()
  return { proofToken, walletAddress }
}

/**
 * @param {string} walletAddress 
 * @param {string} proofToken 
 * @param {Record<string, any>} metadata
 * @returns {Promise<void>}
 */
async function createAccount(walletAddress, proofToken, metadata) {
  const res = await fetch('http://localhost:3000/v1/accounts', {
    method: 'POST',
    headers: { "Content-Type": "application/json", "Api-Key": API_KEY, "Proof-Token": proofToken },
    body: JSON.stringify({ walletAddress, metadata }),
  })
  if (res.status >= 400) {
    const { error } = await res.json()
    console.error(`Error (createAccount):`, error)
    throw new HttpError(res.status, error)
  }
}

/**
 * @param {string} walletAddress 
 * @param {string} proofToken 
 * @returns {Promise<Record<string, any>>} metadata
 */
async function getAccountMetadata(walletAddress, proofToken) {
  const res = await fetch(`http://localhost:3000/v1/accounts/${walletAddress}/metadata`, {
    method: 'GET',
    headers: { "Content-Type": "application/json", "Api-Key": API_KEY, "Proof-Token": proofToken },
  })
  if (res.status >= 400) {
    const { error } = await res.json()
    console.error(`Error (getAccountMetadata):`, error)
    throw new HttpError(res.status, error)
  }
  return (await res.json()).metadata
}

const server = http.createServer(async function (req, res) {
  if (req.method === 'POST') req.body = await readBody()

  const router = {
    GET: {
      '/': () => sendFile('public/index.html', 'text/html'),
      '/index.js': () => sendFile('public/index.js', 'text/javascript'),
      '/index.css': () => sendFile('public/index.css', 'text/css'),
      '/auth/user': async () => {
        const sessionId = getSessionId()
        if (!authenticated(() => sessionId)) return sendJson({ error: 'user not logged in' }, UNAUTHORIZED)
        const session = sessions[sessionId]
        return sendJson({ user: session.user })
      }
    },
    POST: {
      '/auth/challenge': async () => sendJson({ challenge: await issueChallenge(req.body.walletAddress) }),
      '/auth/register': async () => {
        const { walletAddress, challenge, signature, email } = req.body
        const resVerifyChallenge = await verifyChallenge(challenge, signature)
        if (!resVerifyChallenge) return sendJson({ error: 'failed to verify challenge' }, UNAUTHORIZED)
        const { proofToken } = resVerifyChallenge

        const user = { email }
        await createAccount(walletAddress, proofToken, user)

        const newSessionId = crypto.randomBytes(20).toString('hex')
        sessions[newSessionId] = { walletAddress, user }
        res.setHeader('Set-Cookie', cookie.serialize("session", newSessionId))
        sendStatus(NO_CONTENT)
      },
      '/auth/login': async () => {
        const sessionId = getSessionId()
        if (authenticated(() => sessionId)) return sendStatus(NO_CONTENT)

        const { walletAddress, challenge, signature } = req.body
        const resVerifyChallenge = await verifyChallenge(challenge, signature)
        if (!resVerifyChallenge) return sendJson({ error: 'failed to verify challenge' }, UNAUTHORIZED)
        const { proofToken } = resVerifyChallenge

        const user = await getAccountMetadata(walletAddress, proofToken)

        const newSessionId = crypto.randomBytes(20).toString('hex')
        sessions[newSessionId] = { walletAddress, user }
        res.setHeader('Set-Cookie', cookie.serialize("session", newSessionId))
        sendStatus(NO_CONTENT)
      },
    },
    DELETE: {
      '/auth/logout': () => {
        delete sessions[getSessionId()]
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

  /**
   * @returns {string|undefined}
   * */
  function getSessionId() {
    return cookie.parse(req.headers.cookie)["session"]
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

  function authenticated(getSessionIdFn = getSessionId) {
    const sessionId = getSessionIdFn()
    return !!sessionId && !!sessions[sessionId]
  }
})

server.listen(PORT, () => console.log(`Listening on port ${PORT}`))
