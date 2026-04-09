"use client"

import { useEffect, useState } from "react"

interface CommandBlockProps {
  code: string
}

export function CommandBlock({ code }: CommandBlockProps) {
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!copied) return

    const timeout = window.setTimeout(() => setCopied(false), 1500)
    return () => window.clearTimeout(timeout)
  }, [copied])

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code)
      setCopied(true)
    } catch {
      setCopied(false)
    }
  }

  return (
    <div className="relative">
      <button
        type="button"
        onClick={handleCopy}
        className="absolute right-2 top-2 border border-neutral-300 bg-white px-2 py-1 font-mono text-[10px] uppercase tracking-[0.08em] text-neutral-700 transition-colors hover:bg-neutral-100 dark:border-neutral-700 dark:bg-neutral-900 dark:text-neutral-200 dark:hover:bg-neutral-800"
        aria-label="Copy command"
      >
        {copied ? "Copied" : "Copy"}
      </button>
      <pre className="overflow-x-auto border border-neutral-200 bg-neutral-50 p-3 pr-16 font-mono text-sm text-neutral-800 dark:border-neutral-800 dark:bg-neutral-900/60 dark:text-neutral-200">
        {code}
      </pre>
    </div>
  )
}
