const PUBLIC_API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const INTERNAL_API_URL = process.env.API_INTERNAL_URL || PUBLIC_API_URL;

export function getApiBaseUrl(): string {
  return typeof window === "undefined" ? INTERNAL_API_URL : PUBLIC_API_URL;
}
