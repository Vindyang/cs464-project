import { NextResponse } from "next/server";

const GATEWAY_URL = process.env.GATEWAY_URL || "http://localhost:8084";

export async function POST(request: Request) {
  const headers = {
    "X-Request-ID": request.headers.get("X-Request-ID") || "",
  };

  try {
    const gatewayRes = await fetch(`${GATEWAY_URL}/api/v1/files/health/refresh`, {
      method: "POST",
      headers,
      cache: "no-store",
    });

    const gatewayData = await gatewayRes.json().catch(() => ({}));
    return NextResponse.json(gatewayData, { status: gatewayRes.status });
  } catch (error) {
    return NextResponse.json(
      {
        error: "Failed to refresh health",
        details: error instanceof Error ? error.message : "Unknown error",
      },
      { status: 500 }
    );
  }
}
