export const dynamic = 'force-dynamic'

import { registry } from "@/lib/help/error-registry"

export const metadata = {
  title: "Help — Omnishard",
  description:
    "Explanations and recovery steps for every Omnishard error code.",
}

export default function HelpPage() {
  const entries = Object.values(registry)

  return (
    <div className="max-w-3xl space-y-10">
      {/* Header */}
      <div className="border-b pb-4">
        <p className="mb-0.5 font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400">
          Documentation
        </p>
        <h1 className="text-2xl font-semibold tracking-tight">Help & Troubleshooting</h1>
        <p className="mt-2 font-mono text-sm text-neutral-500 dark:text-neutral-400">
          When an action fails, Omnishard shows an error code in the notification. Find your code
          below for the cause and recovery steps.
        </p>
      </div>

      {/* Quick-nav */}
      <nav
        aria-label="Error code index"
        className="border border-neutral-200 bg-neutral-50 px-5 py-4 dark:border-neutral-800 dark:bg-neutral-900/60"
      >
        <p className="mb-3 font-mono text-[11px] uppercase tracking-[0.1em] text-neutral-500">
          Error Codes
        </p>
        <ul className="flex flex-wrap gap-2">
          {entries.map((e) => (
            <li key={e.code}>
              <a
                href={`#${e.docsAnchor}`}
                className="font-mono text-[11px] uppercase tracking-wider text-sky-700 underline underline-offset-2 hover:text-sky-900 dark:text-sky-400 dark:hover:text-sky-200"
              >
                {e.code}
              </a>
            </li>
          ))}
        </ul>
      </nav>

      {/* One section per error code */}
      {entries.map((entry) => (
        <section
          key={entry.code}
          id={entry.docsAnchor}
          className="scroll-mt-6 border border-neutral-200 dark:border-neutral-800"
        >
          {/* Section header */}
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

          {/* Body */}
          <div className="space-y-5 px-5 py-5">
            <div>
              <p className="mb-1 font-mono text-[11px] uppercase tracking-[0.1em] text-neutral-400">
                Why it happens
              </p>
              <p className="font-mono text-sm leading-relaxed text-neutral-700 dark:text-neutral-300">
                {entry.why}
              </p>
            </div>

            <div>
              <p className="mb-2 font-mono text-[11px] uppercase tracking-[0.1em] text-neutral-400">
                What to do
              </p>
              <ol className="space-y-2">
                {entry.steps.map((step, i) => (
                  <li key={i} className="flex gap-3">
                    <span className="shrink-0 font-mono text-[11px] text-neutral-400 dark:text-neutral-500 pt-0.5">
                      {i + 1}.
                    </span>
                    <span className="font-mono text-sm leading-relaxed text-neutral-700 dark:text-neutral-300">
                      {step}
                    </span>
                  </li>
                ))}
              </ol>
            </div>
          </div>
        </section>
      ))}

      {/* Footer note */}
      <p className="border-t pt-5 font-mono text-[11px] text-neutral-400 dark:text-neutral-500">
        If you see an error code not listed here, check the browser console and backend service logs
        for additional context.
      </p>
    </div>
  )
}
