const GATEWAY_URL = process.env.GATEWAY_URL || "http://localhost:8084";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ fileId: string }> }
) {
  const { fileId } = await params;
  const upstream = await fetch(`${GATEWAY_URL}/api/v1/download/${fileId}`, {
    method: "GET",
    headers: {
      "X-Request-ID": request.headers.get("X-Request-ID") || "",
    },
    cache: "no-store",
  });

  if (!upstream.ok || !upstream.body) {
    const data = await upstream.json().catch(() => ({}));
    return Response.json(
      { error: data?.error || "Failed to download file" },
      { status: upstream.status || 500 }
    );
  }

  const headers = new Headers();
  headers.set(
    "Content-Type",
    upstream.headers.get("Content-Type") || "application/octet-stream"
  );
  const disposition = upstream.headers.get("Content-Disposition");
  if (disposition) headers.set("Content-Disposition", disposition);

  return new Response(upstream.body, {
    status: upstream.status,
    headers,
  });
}
