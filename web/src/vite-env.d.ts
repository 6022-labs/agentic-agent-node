/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_APP_KIT_PROJECT_ID?: string
  readonly VITE_ENVIRONMENT?: string
  readonly VITE_GATEWAY_API_BASE_URL?: string
  readonly VITE_SCANNER_BASE_URL?: string
  readonly VITE_TX_GAS_MULTIPLIER?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
