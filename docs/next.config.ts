import { NextConfig } from "next";
import nextra from "nextra";

const withNextra = nextra({
  latex: true,
  defaultShowCopyCode: true,
  contentDirBasePath: "/",
});

const isProd = process.env.NODE_ENV === "production";

const nextConfig: NextConfig = withNextra({
  reactStrictMode: true,
  output: "export",

  ...(isProd && {
    basePath: "/nome-repo",
    assetPrefix: "/nome-repo/",
  }),

  images: {
    unoptimized: true,
  },
});

export default nextConfig;
