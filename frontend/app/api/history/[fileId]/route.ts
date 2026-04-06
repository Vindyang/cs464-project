import { NextResponse } from "next/server";

const GATEWAY_URL = process.env.GATEWAY_URL || "http://localhost:8084";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ fileId: string }> },
) {
  const { fileId } = await params;

  try {
    const gatewayRes = await fetch(`${GATEWAY_URL}/api/v1/history/${fileId}`, {
      method: "GET",
      headers: {
        "X-Request-ID": request.headers.get("X-Request-ID") || "",
      },
      cache: "no-store",
    });

    const gatewayData = await gatewayRes.json().catch(() => ({}));
    return NextResponse.json(gatewayData, { status: gatewayRes.status });
  } catch (error) {
    return NextResponse.json(
      {
        error: "Failed to fetch file history",
        details: error instanceof Error ? error.message : "Unknown error",
      },
      { status: 500 },
    );
  }
}
