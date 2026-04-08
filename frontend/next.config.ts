import type { NextConfig } from "next";
import bundleAnalyzer from "@next/bundle-analyzer";

const withBundleAnalyzer = bundleAnalyzer({
  enabled: process.env.ANALYZE === "true",
});

const nextConfig: NextConfig = {
  output: "standalone",
  outputFileTracingExcludes: {
    "*": [
      "node_modules/typescript/**",
      "node_modules/@img/**",
      "node_modules/sharp/**",
    ],
  },
};

export default withBundleAnalyzer(nextConfig);
