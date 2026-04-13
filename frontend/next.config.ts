import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  env: {
    BACKEND_URL: process.env.BACKEND_URL ?? 'http://localhost:8080',
  },
}

export default nextConfig
