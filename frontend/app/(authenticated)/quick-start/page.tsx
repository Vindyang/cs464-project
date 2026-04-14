import { CommandBlock } from "./CommandBlock"
import { OnThisPage } from "@/components/docs/on-this-page"

const endpoints = [
  { label: "Frontend", url: "http://localhost:3000" },
  { label: "Adapter", url: "http://localhost:8080" },
  { label: "Shardmap", url: "http://localhost:8081" },
  { label: "Orchestrator", url: "http://localhost:8082" },
  { label: "Sharding", url: "http://localhost:8083" },
  { label: "Gateway", url: "http://localhost:8084" },
]

const prerequisites = [
  "64-bit system (macOS/Linux recommended)",
  "Minimum: 2 CPU cores, 6 GB RAM. Recommended: 4 CPU cores, 8 GB RAM.",
  "Docker Desktop (macOS/Windows) or Docker Engine (Linux)",
  "Docker Compose plugin (`docker compose` command)",
  "Internet access for cloud provider OAuth/API calls",
  "Open local ports: `3000`, `8080`, `8081`, `8082`, `8083`, `8084`",
]

const credentialSteps = [
  "Open `http://localhost:3000`",
  "Go to Credentials",
  "Choose provider in Add or Update",
  "Fill fields shown in the form",
  "Click Save Credentials",
  "Go to Providers page and connect/authorize provider if required",
]

const runModes = [
  "Clone + build from source (developer workflow).",
  "Pull the latest published images from GitHub Packages (GHCR).",
]

const setupSections = [
  {
    code: "QS-001",
    title: "Prerequisites",
    sections: [
      {
        label: "Requirements",
        type: "list" as const,
        items: prerequisites,
      },
      {
        label: "Verify Docker",
        type: "code" as const,
        code: `docker --version\ndocker compose version`,
      },
    ],
  },
  {
    code: "QS-002",
    title: "Two Ways to Run Omnishard",
    sections: [
      {
        label: "Supported workflows",
        type: "steps" as const,
        items: runModes,
      },
      {
        label: "Default for this guide",
        type: "text" as const,
        text: "This Quick Start uses the clone + build from source workflow by default.",
      },
      {
        label: "Run from latest GHCR images",
        type: "code" as const,
        code: "curl -L -o docker-compose.yml https://github.com/Vindyang/cs464-project/releases/latest/download/docker-compose.full-microservices.yml",
      },
      {
        label: "Start stack from GHCR compose file",
        type: "code" as const,
        code: "docker compose up -d",
      },
    ],
  },
  {
    code: "QS-003",
    title: "Quick Start (Docker)",
    sections: [
      {
        label: "Run from repo root (`cs464-project/`)",
        type: "code" as const,
        code: "docker compose --profile full up -d --build",
      },
      {
        label: "Endpoints",
        type: "endpoints" as const,
      },
      {
        label: "Stop stack",
        type: "code" as const,
        code: "docker compose --profile full down",
      },
      {
        label: "Reset persisted volumes (clean slate)",
        type: "code" as const,
        code: "docker compose --profile full down -v",
      },
    ],
  },
  {
    code: "QS-004",
    title: "Credential Configuration",
    sections: [
      {
        label: "Provider setup docs",
        type: "text" as const,
        text: "Provider-specific setup links and credential instructions are shown on the Credentials page after you select a provider.",
      },
      {
        label: "Steps",
        type: "steps" as const,
        items: credentialSteps,
      },
    ],
  },
  {
    code: "QS-005",
    title: "Troubleshooting",
    sections: [
      {
        label: "Check container status",
        type: "code" as const,
        code: "docker compose ps",
      },
      {
        label: "Tail logs",
        type: "code" as const,
        code: `docker compose logs -f frontend\ndocker compose logs -f gateway\ndocker compose logs -f adapter\ndocker compose logs -f orchestrator\ndocker compose logs -f shardmap\ndocker compose logs -f sharding`,
      },
    ],
  },
]

export const metadata = {
  title: "Quick Start - Omnishard",
  description:
    "Contributor and operator setup guide for running Omnishard with Docker and configuring provider credentials.",
}

export default function QuickStartPage() {
  const tocItems = setupSections.map((entry) => ({
    id: entry.code.toLowerCase(),
    label: entry.title,
  }))

  return (
    <div className="space-y-10">
      <div className="border-b pb-4">
        <p className="mb-0.5 font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400">
          Documentation
        </p>
        <h1 className="text-2xl font-semibold tracking-tight">Quick Start</h1>
        <p className="mt-2 font-mono text-sm text-neutral-500 dark:text-neutral-400">
          This guide helps contributors and operators run Omnishard with Docker and set up provider
          credentials through the frontend UI.
        </p>
      </div>

      <div className="grid gap-8 xl:grid-cols-[minmax(0,1fr)_260px]">
        <div className="space-y-10">
          {setupSections.map((entry) => (
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

              <div className="space-y-5 px-5 py-5">
                {entry.sections.map((content) => (
                  <div key={content.label}>
                    <p className="mb-2 font-mono text-[11px] uppercase tracking-[0.1em] text-neutral-400">
                      {content.label}
                    </p>

                    {content.type === "text" && (
                      <p className="font-mono text-sm leading-relaxed text-neutral-700 dark:text-neutral-300">
                        {content.text}
                      </p>
                    )}

                    {content.type === "list" && (
                      <ul className="space-y-2">
                        {content.items.map((item) => (
                          <li key={item} className="flex gap-3">
                            <span className="pt-0.5 font-mono text-[11px] text-neutral-400 dark:text-neutral-500">
                              •
                            </span>
                            <span className="font-mono text-sm leading-relaxed text-neutral-700 dark:text-neutral-300">
                              {item}
                            </span>
                          </li>
                        ))}
                      </ul>
                    )}

                    {content.type === "steps" && (
                      <ol className="space-y-2">
                        {content.items.map((step, i) => (
                          <li key={`${content.label}-${i}`} className="flex gap-3">
                            <span className="shrink-0 pt-0.5 font-mono text-[11px] text-neutral-400 dark:text-neutral-500">
                              {i + 1}.
                            </span>
                            <span className="font-mono text-sm leading-relaxed text-neutral-700 dark:text-neutral-300">
                              {step}
                            </span>
                          </li>
                        ))}
                      </ol>
                    )}

                    {content.type === "code" && (
                      <CommandBlock code={content.code} />
                    )}

                    {content.type === "endpoints" && (
                      <ul className="space-y-2">
                        {endpoints.map((endpoint) => (
                          <li key={endpoint.label} className="flex gap-3">
                            <span className="pt-0.5 font-mono text-[11px] text-neutral-400 dark:text-neutral-500">
                              •
                            </span>
                            <span className="font-mono text-sm leading-relaxed text-neutral-700 dark:text-neutral-300">
                              {endpoint.label}:{" "}
                              <a
                                href={endpoint.url}
                                target="_blank"
                                rel="noreferrer"
                                className="underline hover:text-neutral-950 dark:hover:text-neutral-100"
                              >
                                {endpoint.url}
                              </a>
                            </span>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                ))}
              </div>
            </section>
          ))}
        </div>

        <OnThisPage items={tocItems} />
      </div>

      <p className="border-t pt-5 font-mono text-[11px] text-neutral-400 dark:text-neutral-500">
        Keep this page aligned with `DEVDOCS.md` whenever setup commands or provider onboarding
        details change.
      </p>
    </div>
  )
}
