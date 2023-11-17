# Gatekeeper

## Supported/Tested wallets
- Metamask

## Authentication steps
1. Connect wallet
2. Issue challenge
3. Login

### Connect wallet
1. Client connects to wallet
2. Client gets wallet address

### Issue challenge
1. Client asks the server to issue a new challenge (wallet address is sent in the params)
2. Server generates a challenge and sends it to the client -> POST /v1/challenge

### Login
1. Client signs challenge
2. Client sends a login request with wallet address and signed challenge
3. Server verifies signed challenge was signed by wallet address -> POST /v1/challenge/verify
4. Server sets the user as authenticated

## Example
In the example, we try to keep things as simple as possible (we are using a barebones node server).
This way most people should be able to figure out how it could be ported to other languages/frameworks.

## How to run example?
- Run `bin/example`