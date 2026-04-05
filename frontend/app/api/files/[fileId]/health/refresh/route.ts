import { NextResponse } from "next/server";

const ORCHESTRATOR_URL = process.env.ORCHESTRATOR_URL || "http://localhost:8082";

export async function POST(
  request: Request,
  { params }: { params: Promise<{ fileId: string }> }
) {
  const { fileId } = await params;
  const upstream = await fetch(`${ORCHESTRATOR_URL}/api/orchestrator/files/${fileId}/health/refresh`, {
    method: "POST",
    headers: {
      "X-Request-ID": request.headers.get("X-Request-ID") || "",
    },
    cache: "no-store",
  });

  const data = await upstream.json().catch(() => ({}));
  return NextResponse.json(data, { status: upstream.status });
}
