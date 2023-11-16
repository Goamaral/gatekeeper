import Web3SSOFrontend from 'web3-sso/frontend'
import axios from 'axios'

/**
 * @param {HTMLElement} el
 * */
function hideEl (el) {
  el.classList.add('hidden')
}

/**
 * @param {HTMLElement} el
 * */
function showEl (el) {
  el.classList.remove('hidden')
}

window.onload = async () => {
  const connectButtonEl = document.getElementById('connect_button')
  const loginButtonEl = document.getElementById('login_button')
  const logoutButtonEl = document.getElementById('logout_button')
  const errorEl = document.getElementById('error')
  const authUserEl = document.getElementById('auth_user')

  function showError (err) {
    errorEl.innerText = err
  }

  async function fetchAuthUser (visibleError = true) {
    try {
      return (await axios.get('/auth/user')).data.user
    } catch (err) {
      if (visibleError) showError(`${err.response.statusText}: ${err.response.data.error}`)
    }
  }

  const sso = new Web3SSOFrontend()
  await sso.init()

  if (!sso.connected) {
    showEl(connectButtonEl)
  } else {
    const user = await fetchAuthUser(false)
    if (!user) {
      showEl(loginButtonEl)
    } else {
      authUserEl.innerText = JSON.stringify(user)
      showEl(logoutButtonEl)
    }
  }

  // Connect to wallet
  connectButtonEl.onclick = async () => {
    await sso.connectWallet()

    hideEl(connectButtonEl)
    showEl(loginButtonEl)
  }

  // Request challenge, sign it, and authenticate
  loginButtonEl.onclick = async () => {
    try {
      await sso.login()
      showError('')
    } catch (err) {
      showError(`${err.response.statusText}: ${err.response.data.error}`)
      return
    }

    const user = await fetchAuthUser()
    authUserEl.innerText = JSON.stringify(user)
    hideEl(loginButtonEl)
    showEl(logoutButtonEl)
  }

  // Logout
  logoutButtonEl.onclick = async () => {
    try {
      await axios.post('/auth/logout')
      showError('')
    } catch (err) {
      showError(`${err.response.statusText}: ${err.response.data.error}`)
      return
    }

    showEl(loginButtonEl)
    hideEl(logoutButtonEl)
    authUserEl.innerText = ''
  }
}
