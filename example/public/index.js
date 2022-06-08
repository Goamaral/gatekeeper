window.onload = () => {
  const connectButtonEl = document.getElementById("connect_button")
  const loginButtonEl = document.getElementById("login_button")
  const getAuthUserButtonEl = document.getElementById("get_auth_user")
  const errorEl = document.getElementById("error")
  const authUserEl = document.getElementById("auth_user")

  if (!window.ethereum) {
    window.alert("Metamask not installed")
  }

  // TODO: Skip connect if already connected
  loginButtonEl.style.display = "none"

  let signer, walletAddress

  // Connect to wallet
  connectButtonEl.onclick = async () => {
    const provider = new ethers.providers.Web3Provider(window.ethereum)
    provider.send("eth_requestAccounts", [])
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
      errorEl.innerText = `${err.response.statusText}: ${err.response.data.error}`
    }


    loginButtonEl.style.display = "none"
  }

  // Get authenticated user
  getAuthUserButtonEl.onclick = async () => {
    try {
      const { user } = (await axios.get("/auth/user")).data
      authUserEl.innerText = JSON.stringify(user)
    } catch (err) {
      errorEl.innerText = `${err.response.statusText}: ${err.response.data.error}`
    }
  }
}