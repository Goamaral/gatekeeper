window.onload = () => {
  const challengeInputEl = document.getElementById("challenge_input")
  const connectButtonEl = document.getElementById("connect_button")
  const loginButtonEl = document.getElementById("login_button")

  if (!window.ethereum) {
    window.alert("Metamask not installed")
  }

  // TODO: Skip connect if already connected
  loginButtonEl.style.display = "none"

  const store = {}

  connectButtonEl.onclick = async () => {
    store.account = ethereum.request({ method: 'eth_requestAccounts' })[0]
    connectButtonEl.style.display = "none"
    loginButtonEl.style.display = "inline-block"

    const provider = new ethers.providers.Web3Provider(window.ethereum)
    store.signer = provider.getSigner()
  }

  loginButtonEl.onclick = async () => {
    const signedChallenge = await store.signer.signMessage(challengeInputEl.value)
    console.log(signedChallenge)
    // TODO: Show signed challenge in the browser
    // TODO: Issue challenge
    // TODO: Sign challenge
    // TODO: Show secret
  }
}