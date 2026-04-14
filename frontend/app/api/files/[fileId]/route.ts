import { NextResponse } from "next/server";
import { revalidatePath } from "next/cache";

const GATEWAY_URL = process.env.GATEWAY_URL || "http://localhost:8084";

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ fileId: string }> }
) {
  const { fileId } = await params;
  const url = new URL(`${GATEWAY_URL}/api/v1/files/${fileId}`);
  const deleteShards = new URL(request.url).searchParams.get("delete_shards");
  if (deleteShards === "true") {
    url.searchParams.set("delete_shards", "true");
  }

  const upstream = await fetch(url.toString(), {
    method: "DELETE",
    headers: {
      "X-Request-ID": request.headers.get("X-Request-ID") || "",
    },
    cache: "no-store",
  });

  if (upstream.status === 204) {
    revalidatePath("/files");
    revalidatePath("/dashboard");
    return new NextResponse(null, { status: 204 });
  }

  const data = await upstream.json().catch(() => ({}));
  if (upstream.ok) {
    revalidatePath("/files");
    revalidatePath("/dashboard");
    return NextResponse.json(data, { status: upstream.status });
  }

  return NextResponse.json({ error: data?.error || "Failed to delete file" }, { status: upstream.status || 500 });
}
