import type { NextConfig } from "next";

const apiURL = process.env.GAME_API_URL ?? "http://127.0.0.1:8080";
const staticExport = process.env.NEXT_EXPORT === "1";

const config: NextConfig = {
  devIndicators: false,
  output: staticExport ? "export" : undefined,
  trailingSlash: true,
};

if (!staticExport) {
  config.rewrites = async () => [{ source: "/api/:path*", destination: `${apiURL}/:path*` }];
}

export default config;
