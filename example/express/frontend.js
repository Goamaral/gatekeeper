import Web3SSOFrontend from 'web3-sso/frontend'
import axios from 'axios'

window.onload = async () => {
  const connectButtonEl = document.getElementById('connect_button')
  const loginButtonEl = document.getElementById('login_button')
  const errorEl = document.getElementById('error')
  const authUserEl = document.getElementById('auth_user')

  function showError (err) {
    errorEl.innerText = err
  }

  async function fetchAuthUser () {
    try {
      return (await axios.get('/auth/user')).data.user
    } catch (err) {
      showError(`${err.response.statusText}: ${err.response.data.error}`)
    }
  }

  const sso = new Web3SSOFrontend()
  await sso.init()

  connectButtonEl.style.display = 'none'
  loginButtonEl.style.display = 'none'

  if (!sso.connected) {
    connectButtonEl.style.display = 'inline-block'
  } else {
    const user = await fetchAuthUser()
    if (!user) {
      loginButtonEl.style.display = 'inline-block'
    } else {
      authUserEl.innerText = JSON.stringify(user)
    }
  }

  // Connect to wallet
  connectButtonEl.onclick = async () => {
    await sso.connectWallet()

    connectButtonEl.style.display = 'none'
    loginButtonEl.style.display = 'inline-block'
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

    loginButtonEl.style.display = 'none'

    const user = await fetchAuthUser()
    if (user) authUserEl.innerText = JSON.stringify(user)
  }
}
