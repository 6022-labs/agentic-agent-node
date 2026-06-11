# Agent node console

Web console for a single **agentic-agent-node**. It talks only to the node's
public **gateway host** (the sole externally reachable surface). Built with
React + TypeScript + Vite + Tailwind v4 + shadcn/ui, Reown AppKit + ethers for
wallets, and axios for the gateway API.

## Features

- **SIWE auth** — connect a wallet (Reown AppKit), then `GET /auth/challenge` →
  `personal_sign(nonce)` → `POST /auth/verify` → JWT (stored in `localStorage`,
  sent as `Bearer`). Anyone may sign in while the agent is unminted; once minted
  the gateway only authorizes the NFT owner / agent wallet (enforced server-side).
- **LLM configuration** — `GET/POST/DELETE /config/llm` (one connection per node).
- **MCP configuration** — `GET/POST/DELETE /config/mcp-connections`.
- **Identity (mint + ENS)** — save a draft, then either:
  - **Self-register** with the node wallet (`POST /agent/registration/self`), or
  - **Browser wallet** — `GET .../mint-info` → call `AgentCollection.mint` /
    `createMintProposal` via ethers → `POST .../mint/check`, then
    `GET .../ens-info` → `AgentEnsRegistry.initAgentTexts` → `POST .../ens/check`.
- **Playground** — chat through `POST /responses`.

## Setup

```bash
npm install
cp .env.example .env   # then edit values
npm run dev
```

### Environment

All vars are optional and fall back to Polygon Amoy defaults (see
`src/configs/env.ts`). Create a `.env` to override:

| Var | Purpose | Default |
| --- | --- | --- |
| `VITE_APP_KIT_PROJECT_ID` | Reown AppKit project id | a shared dev id |
| `VITE_GATEWAY_API_BASE_URL` | Gateway host base URL | `http://localhost:5000` |
| `VITE_SCANNER_BASE_URL` | Block explorer base URL | Amoy polygonscan |
| `VITE_ENVIRONMENT` | `development` / `production` | `development` |

> The ABIs in `src/abi/` are extracted from the `agentic-smart-contracts`
> artifacts. The `mint` signature here is the agent-node direct mint
> `mint(to, role, name, wallets, images[], attributes[], cloneOf)` — no
> signer/nonce/signature.

## Scripts

```bash
npm run dev        # vite dev server
npm run build      # tsc -b && vite build
npm run typecheck  # tsc --noEmit
npm run lint       # eslint
```
