"use client"

import { useEffect, useState } from "react"

interface OnThisPageItem {
  id: string
  label: string
}

interface OnThisPageProps {
  items: OnThisPageItem[]
  title?: string
}

export function OnThisPage({ items, title = "On this page" }: OnThisPageProps) {
  const [activeId, setActiveId] = useState(items[0]?.id ?? "")

  useEffect(() => {
    if (!items.length) return

    const getScrollParent = (element: HTMLElement | null): HTMLElement | Window => {
      let current = element?.parentElement ?? null

      while (current) {
        const styles = window.getComputedStyle(current)
        const overflowY = styles.overflowY
        if (overflowY === "auto" || overflowY === "scroll") {
          return current
        }
        current = current.parentElement
      }

      return window
    }

    const marker = document.getElementById(items[0].id)
    const scrollTarget = getScrollParent(marker)

    const updateActiveId = () => {
      const threshold =
        scrollTarget === window
          ? 140
          : (scrollTarget as HTMLElement).getBoundingClientRect().top + 120

      const atBottom =
        scrollTarget === window
          ? window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 4
          : (scrollTarget as HTMLElement).scrollTop +
              (scrollTarget as HTMLElement).clientHeight >=
            (scrollTarget as HTMLElement).scrollHeight - 4

      if (atBottom) {
        setActiveId(items[items.length - 1]?.id ?? "")
        return
      }

      let current = items[0]?.id ?? ""

      for (const item of items) {
        const element = document.getElementById(item.id)
        if (!element) continue

        if (element.getBoundingClientRect().top <= threshold) {
          current = item.id
        } else {
          break
        }
      }

      setActiveId(current)
    }

    let rafId = 0
    const onScrollOrResize = () => {
      cancelAnimationFrame(rafId)
      rafId = requestAnimationFrame(updateActiveId)
    }

    updateActiveId()
    if (scrollTarget === window) {
      window.addEventListener("scroll", onScrollOrResize, { passive: true })
    } else {
      ;(scrollTarget as HTMLElement).addEventListener("scroll", onScrollOrResize, {
        passive: true,
      })
    }
    window.addEventListener("resize", onScrollOrResize)

    return () => {
      cancelAnimationFrame(rafId)
      if (scrollTarget === window) {
        window.removeEventListener("scroll", onScrollOrResize)
      } else {
        ;(scrollTarget as HTMLElement).removeEventListener("scroll", onScrollOrResize)
      }
      window.removeEventListener("resize", onScrollOrResize)
    }
  }, [items])

  return (
    <aside className="hidden xl:block">
      <div className="sticky top-20 border-l border-neutral-200 pl-6 dark:border-neutral-800">
        <p className="mb-4 text-lg font-medium text-neutral-900 dark:text-neutral-100">
          {title}
        </p>
        <ul className="space-y-3">
          {items.map((item) => {
            const isActive = activeId === item.id
            return (
              <li key={item.id}>
                <a
                  href={`#${item.id}`}
                  className={
                    isActive
                      ? "text-sm font-medium leading-relaxed transition-colors"
                      : "text-sm leading-relaxed text-neutral-500 transition-colors hover:text-sky-500 dark:text-neutral-400 dark:hover:text-sky-400"
                  }
                  style={isActive ? { color: "rgb(14 165 233)" } : undefined}
                  aria-current={isActive ? "true" : undefined}
                >
                  {item.label}
                </a>
              </li>
            )
          })}
        </ul>
      </div>
    </aside>
  )
}
