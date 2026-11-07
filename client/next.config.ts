import type { NextConfig } from "next";

const apiURL = process.env.GAME_API_URL ?? "http://127.0.0.1:8080";

const config: NextConfig = {
  devIndicators: false,
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${apiURL}/:path*` }];
  },
};

export default config;
