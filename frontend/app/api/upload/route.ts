import { NextResponse } from "next/server";

const GATEWAY_URL = process.env.GATEWAY_URL || "http://localhost:8084";

export async function POST(request: Request) {
  try {
    const form = await request.formData();
    const file = form.get("file");
    const k = String(form.get("k") || "4");
    const n = String(form.get("n") || "6");

    if (!(file instanceof File)) {
      return NextResponse.json({ error: "Missing file field" }, { status: 400 });
    }

    const upstreamForm = new FormData();
    upstreamForm.append("file", file, file.name);
    upstreamForm.append("k", k);
    upstreamForm.append("n", n);

    const upstreamRes = await fetch(`${GATEWAY_URL}/api/v1/upload`, {
      method: "POST",
      body: upstreamForm,
      cache: "no-store",
    });

    const data = await upstreamRes.json().catch(() => ({}));
    return NextResponse.json(data, { status: upstreamRes.status });
  } catch (error) {
    return NextResponse.json(
      {
        error: "Failed to upload file",
        code: "UNKNOWN_ERROR",
        details: error instanceof Error ? error.message : "Unknown error",
      },
      { status: 500 }
    );
  }
}
