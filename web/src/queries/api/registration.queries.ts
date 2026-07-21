import axios from "axios"

import type {
  AgentRegistration,
  EnsInfo,
  MintInfo,
  RegistrationStatus,
  SaveAgentRegistrationRequest,
} from "@/models/agent-registration"

// GET /agent/registration — 404 draft_not_found when nothing saved yet.
export const getAgentRegistration = async () => {
  return await axios.get<AgentRegistration>("/agent/registration")
}

// POST /agent/registration — persist the owner-filled draft (multipart upload).
export const saveAgentRegistration = async (
  request: SaveAgentRegistrationRequest
) => {
  const form = new FormData()
  form.append("chainId", request.chainId)
  form.append("collectionAddress", request.collectionAddress)
  form.append("name", request.name)
  if (request.cloneOf != null) {
    form.append("cloneOf", String(request.cloneOf))
  }
  // Required attributes travel inside the attributes JSON.
  form.append(
    "attributes",
    JSON.stringify({ ...request.attributes, role: request.role })
  )
  // Omit on re-edit to keep the stored image; only sent when a new one is picked.
  if (request.image) {
    form.append("image", request.image)
  }
  return await axios.post<AgentRegistration>("/agent/registration", form)
}

// GET /agent/registration/mint-info — mint() calldata + moderator state.
export const getAgentMintInfo = async () => {
  return await axios.get<MintInfo>("/agent/registration/mint-info")
}

// POST /agent/registration/mint/check — reconcile mint status from chain.
export const checkAgentMint = async () => {
  return await axios.post<RegistrationStatus>("/agent/registration/mint/check")
}

// GET /agent/registration/ens-info — initAgentTexts() calldata (409 not_minted).
export const getAgentEnsInfo = async () => {
  return await axios.get<EnsInfo>("/agent/registration/ens-info")
}

// POST /agent/registration/ens/check — reconcile ENS status from chain.
export const checkAgentEns = async () => {
  return await axios.post<RegistrationStatus>("/agent/registration/ens/check")
}

// POST /agent/registration/self — node wallet mints + ENS-registers.
export const selfRegisterAgent = async () => {
  return await axios.post<RegistrationStatus>("/agent/registration/self")
}
