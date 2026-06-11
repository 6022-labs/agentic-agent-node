import axios from "axios"

export const initAxiosDefaults = () => {
  // Point at the agent-node gateway (the only public surface). Falls back to the
  // legacy var name so an existing .env keeps working.
  // Note: no blanket Content-Type default — axios picks application/json for
  // plain objects and multipart/form-data for FormData bodies; forcing json
  // here would make axios serialize FormData (drafts, image upload) to JSON.
  axios.defaults.baseURL = import.meta.env.VITE_GATEWAY_API_BASE_URL

  axios.interceptors.request.use((config) => {
    const jwt = localStorage.getItem("jwt")
    if (jwt && !config.headers?.Authorization) {
      config.headers = config.headers ?? {}
      config.headers.Authorization = `Bearer ${jwt}`
    }
    return config
  })
}
