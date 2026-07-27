import { NextConfig } from "next";
import nextra from "nextra";

const withNextra = nextra({
  latex: true,
  defaultShowCopyCode: true,
  contentDirBasePath: "/",
});

const isGithubPages = process.env.NEXT_PUBLIC_GITHUB_PAGES === "true";

const nextConfig: NextConfig = withNextra({
  reactStrictMode: true,
  output: "export",

  ...(isGithubPages && {
    basePath: "/LWN-Simulator",
    assetPrefix: "/LWN-Simulator/",
  }),

  images: {
    unoptimized: true,
  },
});

export default nextConfig;
