import { toast } from "sonner"
import { lookupHelp } from "./error-registry"

type ApiError = {
  error?: string
  code?: string
  details?: string
  message?: string
}

/**
 * Shows an error toast with contextual help looked up from the error registry.
 * Pass the parsed JSON error body from any API response, or a plain Error/string.
 * Falls back gracefully when code is missing or unknown.
 */
export function helpToast(err: ApiError | Error | string | unknown): void {
  const apiErr = normaliseError(err)
  const entry = lookupHelp(apiErr.code)

  const description =
    entry.steps.length > 0
      ? `${entry.steps[0]}${entry.steps.length > 1 ? ` (+${entry.steps.length - 1} more steps)` : ""}`
      : undefined

  toast.error(entry.title, {
    description,
    action: {
      label: "Learn more →",
      onClick: () => {
        window.location.href = `/help#${entry.docsAnchor}`
      },
    },
    duration: 6000,
  })
}

function normaliseError(err: unknown): ApiError {
  if (typeof err === "string") {
    return { error: err }
  }
  if (err instanceof Error) {
    return { error: err.message }
  }
  if (err && typeof err === "object") {
    return err as ApiError
  }
  return {}
}
