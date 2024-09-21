import Gatekeeper, { MetamaskNotInstalledError } from 'gatekeeper'
import axios from 'axios'

const CONNECT_MODE = 1
const AUTH_MODE = 2
const LOGOUT_MODE = 3

const CONNECT_WALLET_MSG = 'Connect wallet'
const LOGIN_MSG = 'Login'
const LOGOUT_MSG = 'Logout'

window.onload = async () => {
  const userEl = document.getElementById('user')
  const errorEl = document.getElementById('error')
  const errorMsgEl = document.getElementById('error-msg')
  const firstButtonEl = document.getElementById('first-button')
  const registerEl = document.getElementById('register')
  const registerEmailEl = document.getElementById('register-email')
  const registerButtonEl = document.getElementById('register-button')

  let gatekeeper
  try {
    gatekeeper = new Gatekeeper()
    await gatekeeper.init()
  } catch (err) {
    if (err instanceof MetamaskNotInstalledError) {
      showError(err)
      firstButtonEl.innerText = 'Please install Metamask'
      return
    } else {
      throw err
    }
  }

  if (!gatekeeper.connected) {
    setButtonMode(CONNECT_MODE)
  } else {
    const user = await fetchAuthUser(false)
    if (!user) {
      setButtonMode(AUTH_MODE)
    } else {
      setUser(user)
      setButtonMode(LOGOUT_MODE)
    }
  }

  function showError(err) {
    if (err instanceof Error) {
      errorMsgEl.innerText = err.message
    } else {
      errorMsgEl.innerText = err
    }
    errorEl.classList.remove('hidden')
  }

  async function fetchAuthUser(visibleError = true) {
    try {
      return (await axios.get('/auth/user')).data.user
    } catch (err) {
      if (visibleError) showError(`${err.response.statusText}: ${err.response.data.error}`)
    }
  }

  function setUser(user) {
    userEl.innerText = user ? JSON.stringify(user) : 'Not authenticated'
  }

  async function setButtonMode(mode) {
    errorEl.classList.add('hidden')
    registerEl.classList.add('hidden')
    registerButtonEl.classList.add('hidden')

    switch (mode) {
      case CONNECT_MODE:
        setUser(undefined)
        firstButtonEl.innerText = CONNECT_WALLET_MSG
        firstButtonEl.onclick = async () => {
          await gatekeeper.connectWallet()
          setButtonMode(AUTH_MODE)
        }
        break

      case AUTH_MODE:
        setUser(undefined)
        firstButtonEl.innerText = LOGIN_MSG
        firstButtonEl.onclick = async () => {
          try {
            await gatekeeper.login()
          } catch (err) {
            return showError(err)
          }
          setButtonMode(LOGOUT_MODE)
        }
        registerEl.classList.remove('hidden')
        registerButtonEl.classList.remove('hidden')
        registerButtonEl.onclick = async () => {
          try {
            await gatekeeper.register(registerEmailEl.value)
          } catch (err) {
            return showError(err)
          }
          setButtonMode(LOGOUT_MODE)
        }
        break

      case LOGOUT_MODE:
        setUser(await fetchAuthUser())
        firstButtonEl.innerText = LOGOUT_MSG
        firstButtonEl.onclick = async () => {
          try {
            await gatekeeper.logout()
          } catch (err) {
            return showError(err)
          }
          setButtonMode(AUTH_MODE)
        }
        break
    }
  }
}
