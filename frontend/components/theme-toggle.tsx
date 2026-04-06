"use client";

import { useSyncExternalStore } from "react";
import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";

export function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme();
  const mounted = useSyncExternalStore(
    () => () => {},
    () => true,
    () => false,
  );

  if (!mounted) {
    return (
      <button
        type="button"
        disabled
        className="flex w-full items-center justify-between border px-3 py-2 font-mono text-[11px] uppercase tracking-[0.08em] text-neutral-600 transition-colors dark:text-neutral-300"
        aria-label="Theme mode toggle"
      >
        <span>Theme Mode</span>
        <Sun className="h-3.5 w-3.5 opacity-60" />
      </button>
    );
  }

  const isDark = resolvedTheme === "dark";

  return (
    <button
      type="button"
      onClick={() => setTheme(isDark ? "light" : "dark")}
      className="flex w-full items-center justify-between border px-3 py-2 font-mono text-[11px] uppercase tracking-[0.08em] text-neutral-600 transition-colors hover:bg-neutral-100 hover:text-neutral-900 dark:text-neutral-300 dark:hover:bg-neutral-900 dark:hover:text-white"
      aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
    >
      <span>{isDark ? "Light Mode" : "Dark Mode"}</span>
      {isDark ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
    </button>
  );
}
