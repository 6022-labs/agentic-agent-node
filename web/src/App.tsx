import React from "react"
import { Navigate, Route, Routes } from "react-router-dom"
import { createAppKit } from "@reown/appkit/react"
import { EthersAdapter } from "@reown/appkit-adapter-ethers"
import { type AppKitNetwork, polygonAmoy } from "@reown/appkit/networks"

import { config } from "@/configs/env"
import { getAppConfig } from "@/configs/app-config"
import { networkByChainId } from "@/configs/networks"
import AppKitThemeSync from "@/components/appkit-theme-sync"
import LoginPage from "@/pages/login"
import DashboardLayout from "@/components/dashboard/dashboard-layout"
import DashboardIdentityPage from "@/pages/dashboard-identity"
import DashboardLlmPage from "@/pages/dashboard-llm"
import DashboardMcpPage from "@/pages/dashboard-mcp"
import DashboardMemoryPage from "@/pages/dashboard-memory"
import DashboardPaymentsPage from "@/pages/dashboard-payments"
import DashboardPlaygroundPage from "@/pages/dashboard-playground"

const metadata = {
  name: "Agentic Agent Node",
  description: "Console for an agentic agent node.",
  url: window.location.origin,
  icons: ["/vite.svg"],
}

function App() {
  React.useMemo(() => {
    // App renders under ConfigGate, so the runtime config is loaded by now.
    // AppKit is initialized with every chain the node serves.
    const configured = Object.keys(getAppConfig().chains)
      .map((id) => networkByChainId[Number(id)])
      .filter((n): n is AppKitNetwork => Boolean(n))
    const networks: [AppKitNetwork, ...AppKitNetwork[]] =
      configured.length > 0
        ? (configured as [AppKitNetwork, ...AppKitNetwork[]])
        : [polygonAmoy]
    createAppKit({
      adapters: [new EthersAdapter()],
      networks,
      metadata,
      projectId: config.appKitProjectId,
      features: {
        email: false,
        socials: false,
        analytics: false,
        connectMethodsOrder: ["wallet"],
      },
    })
  }, [])

  return (
    <>
      <AppKitThemeSync />
      <Routes>
        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/dashboard" element={<DashboardLayout />}>
          <Route
            index
            element={<Navigate to="/dashboard/identity" replace />}
          />
          <Route
            path="config"
            element={<Navigate to="/dashboard/identity" replace />}
          />
          <Route path="identity" element={<DashboardIdentityPage />} />
          <Route path="llm" element={<DashboardLlmPage />} />
          <Route path="mcp" element={<DashboardMcpPage />} />
          <Route path="memory" element={<DashboardMemoryPage />} />
          <Route path="payments" element={<DashboardPaymentsPage />} />
          <Route path="playground" element={<DashboardPlaygroundPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </>
  )
}

export default App
