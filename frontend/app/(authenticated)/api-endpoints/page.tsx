import { OnThisPage } from "@/components/docs/on-this-page";

const endpointDocs = [
  {
    code: "API-001",
    title: "Available Endpoints",
    description: "",
  },
];

async function isReachable(url: string): Promise<boolean> {
  if (!url) return false;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 1200);

  try {
    const response = await fetch(url, {
      method: "GET",
      cache: "no-store",
      signal: controller.signal,
    });
    return response.ok;
  } catch {
    return false;
  } finally {
    clearTimeout(timeoutId);
  }
}

async function resolveDocsUrl(): Promise<string> {
  const gatewayUrl = process.env.GATEWAY_URL ?? "";
  const internalApiUrl = process.env.API_INTERNAL_URL ?? "";

  if (await isReachable(`${gatewayUrl}/api/v1/docs`)) {
    const publicGatewayBase = gatewayUrl.includes(":8084")
      ? "http://localhost:8084"
      : "http://localhost:8080";
    return `${publicGatewayBase}/api/v1/docs`;
  }

  if (await isReachable(`${internalApiUrl}/api/v1/docs`)) {
    return "http://localhost:8080/api/v1/docs";
  }

  return "http://localhost:8080/api/v1/docs";
}

export const metadata = {
  title: "API Endpoints - Omnishard",
  description: "Quick links to Omnishard live endpoint documentation.",
};

export default async function ApiEndpointsPage() {
  const docsUrl = await resolveDocsUrl();
  const tocItems = endpointDocs.map((entry) => ({
    id: entry.code.toLowerCase(),
    label: entry.title,
  }));

  return (
    <div className="space-y-10">
      <div className="border-b pb-4">
        <p className="mb-0.5 font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400">
          Documentation
        </p>
        <h1 className="text-2xl font-semibold tracking-tight">API Endpoints</h1>
        <p className="mt-2 font-mono text-sm text-neutral-500 dark:text-neutral-400">
          This page points to the live backend docs so you can see currently
          available endpoints without duplicating endpoint definitions in the
          frontend.
        </p>
      </div>

      <div className="grid gap-8 xl:grid-cols-[minmax(0,1fr)_260px]">
        <div className="space-y-10">
          {endpointDocs.map((entry) => (
            <section
              key={entry.code}
              id={entry.code.toLowerCase()}
              className="scroll-mt-6 border border-neutral-200 dark:border-neutral-800"
            >
              <div className="border-b border-neutral-200 bg-neutral-50 px-5 py-4 dark:border-neutral-800 dark:bg-neutral-900/60">
                <div className="flex flex-wrap items-baseline gap-3">
                  <span className="font-mono text-[11px] uppercase tracking-[0.12em] text-neutral-400">
                    {entry.code}
                  </span>
                  <h2 className="text-base font-semibold text-neutral-900 dark:text-neutral-100">
                    {entry.title}
                  </h2>
                </div>
              </div>

              <div className="space-y-4 px-5 py-5">
                <p className="font-mono text-sm leading-relaxed text-neutral-700 dark:text-neutral-300">
                  {entry.description}
                </p>
                <a
                  href={docsUrl}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex border border-sky-600 bg-sky-600 px-4 py-2 font-mono text-[11px] uppercase tracking-[0.08em] text-white transition-colors hover:bg-sky-700"
                >
                  Open Endpoint Docs
                </a>
                <p className="font-mono text-xs text-neutral-500 dark:text-neutral-400">
                  Resolved URL:{" "}
                  <a
                    href={docsUrl}
                    target="_blank"
                    rel="noreferrer"
                    className="underline hover:text-neutral-950 dark:hover:text-neutral-100"
                  >
                    {docsUrl}
                  </a>
                </p>
              </div>
            </section>
          ))}
        </div>

        <OnThisPage items={tocItems} />
      </div>

      <p className="border-t pt-5 font-mono text-[11px] text-neutral-400 dark:text-neutral-500">
        The URL is selected automatically from the running backend flavor and
        defaults to monolith docs if detection is unavailable.
      </p>
    </div>
  );
}
