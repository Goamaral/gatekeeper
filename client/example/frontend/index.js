import Gatekeeper, { MetamaskNotInstalledError } from 'gatekeeper'
import axios from 'axios'

const CONNECT_MODE = 1
const LOGIN_MODE = 2
const LOGOUT_MODE = 3

const CONNECT_WALLET_MSG = 'Connect wallet'
const LOGIN_MSG = 'Login'
const LOGOUT_MSG = 'Logout'

window.onload = async () => {
  const buttonEl = document.getElementById('button')
  const userEl = document.getElementById('user')
  const errorEl = document.getElementById('error')
  const errorMsgEl = document.getElementById('error_msg')

  let gatekeeper
  try {
    gatekeeper = new Gatekeeper()
    await gatekeeper.init()
  } catch (err) {
    if (err instanceof MetamaskNotInstalledError) {
      showError(err)
      buttonEl.innerText = 'Please install Metamask'
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
      setButtonMode(LOGIN_MODE)
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

    switch (mode) {
      case CONNECT_MODE:
        setUser(undefined)
        buttonEl.innerText = CONNECT_WALLET_MSG
        buttonEl.onclick = async () => {
          await gatekeeper.connectWallet()
          setButtonMode(LOGIN_MODE)
        }
        break

      case LOGIN_MODE:
        setUser(undefined)
        buttonEl.innerText = LOGIN_MSG
        buttonEl.onclick = async () => {
          try {
            await gatekeeper.login()
          } catch (err) {
            return showError(err)
          }
          setButtonMode(LOGOUT_MODE)
        }
        break

      case LOGOUT_MODE:
        setUser(await fetchAuthUser())
        buttonEl.innerText = LOGOUT_MSG
        buttonEl.onclick = async () => {
          try {
            await gatekeeper.logout()
          } catch (err) {
            return showError(err)
          }
          setButtonMode(LOGIN_MODE)
        }
        break
    }
  }
}
