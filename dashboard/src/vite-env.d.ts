/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE?: string
  readonly VITE_WELL_ID?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
