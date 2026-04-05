import { NextResponse } from "next/server";

const GATEWAY_URL = process.env.GATEWAY_URL || "http://localhost:8084";
const ORCHESTRATOR_URL = process.env.ORCHESTRATOR_URL || "http://localhost:8082";

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

    // If gateway route isn't available yet (e.g. stale gateway config),
    // fallback directly to orchestrator so refresh still works.
    if (gatewayRes.status === 403 || gatewayRes.status === 404 || gatewayRes.status === 405) {
      const orchestratorRes = await fetch(`${ORCHESTRATOR_URL}/api/orchestrator/files/health/refresh`, {
        method: "POST",
        headers,
        cache: "no-store",
      });
      const orchestratorData = await orchestratorRes.json().catch(() => ({}));
      return NextResponse.json(orchestratorData, { status: orchestratorRes.status });
    }

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
