import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const API_URL = process.env.API_INTERNAL_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function hasConfiguredCredentials(): Promise<boolean> {
  try {
    const res = await fetch(`${API_URL}/api/credentials/status`, {
      method: "GET",
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
    });
    if (!res.ok) return false;
    const data = (await res.json()) as { configured?: boolean };
    return Boolean(data.configured);
  } catch {
    return false;
  }
}

export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Only protect app pages that should require configured credentials.
  const protectedRoutes = ["/dashboard", "/files", "/history", "/providers", "/nodes", "/settings"];
  const isProtected = protectedRoutes.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`),
  );

  if (!isProtected) {
    return NextResponse.next();
  }

  const configured = await hasConfiguredCredentials();
  if (!configured) {
    const url = request.nextUrl.clone();
    url.pathname = "/credentials";
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*", "/files/:path*", "/history/:path*", "/providers/:path*", "/nodes/:path*", "/settings/:path*"],
};
