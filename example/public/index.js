window.onload = async () => {
  const connectButtonEl = document.getElementById("connect_button")
  const loginButtonEl = document.getElementById("login_button")
  const errorEl = document.getElementById("error")
  const authUserEl = document.getElementById("auth_user")

  connectButtonEl.style.display = "none"
  loginButtonEl.style.display = "none"

  let user = null
  let signer, walletAddress
  
  const fetchAuthUser = async (showError = true) => {
    try {
      user = (await axios.get("/auth/user")).data.user
    } catch (err) {
      if (showError) errorEl.innerText = `${err.response.statusText}: ${err.response.data.error}`
    }
  }
  
  if (!window.ethereum) {
    window.alert("Metamask not installed")
  }
  
  const provider = new ethers.providers.Web3Provider(window.ethereum)

  const alreadyConnected = (await provider.listAccounts()).length != 0
  if (!alreadyConnected) {
    connectButtonEl.style.display = "inline-block"
  } else {
    signer = provider.getSigner()
    walletAddress = await signer.getAddress()
    await fetchAuthUser(false)

    if (!user) {
      loginButtonEl.style.display = "inline-block"
    } else {
      authUserEl.innerText = JSON.stringify(user)
    }
  }

  // Connect to wallet
  connectButtonEl.onclick = async () => {
    await provider.send("eth_requestAccounts", [])
    signer = provider.getSigner()
    walletAddress = await signer.getAddress()

    connectButtonEl.style.display = "none"
    loginButtonEl.style.display = "inline-block"
  }

  // Request challenge, sign it, and authenticate
  loginButtonEl.onclick = async () => {
    try {
      const { challenge } = (await axios.post("/auth/challenge", { wallet_address: walletAddress })).data
      const signedChallenge = await signer.signMessage(challenge)
      await axios.post("/auth/login", { wallet_address: walletAddress, signed_challenge: signedChallenge })
    } catch (err) {
      console.error(err)
      errorEl.innerText = `${err.response.statusText}: ${err.response.data.error}`
    }

    loginButtonEl.style.display = "none"

    await fetchAuthUser()
    if (user) authUserEl.innerText = JSON.stringify(user)
  }
}