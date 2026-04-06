"use client";

import { startTransition, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";

const FIVE_MINUTES_MS = 5 * 60 * 1000;

export function HealthPollingController() {
  const router = useRouter();
  const inFlightRef = useRef(false);
  const lastRefreshRef = useRef(Date.now());

  useEffect(() => {
    async function refreshHealth() {
      if (document.visibilityState !== "visible" || inFlightRef.current) {
        return;
      }

      inFlightRef.current = true;
      try {
        await fetch("/api/files/health/refresh", {
          method: "POST",
          cache: "no-store",
        });
        lastRefreshRef.current = Date.now();
        startTransition(() => {
          router.refresh();
        });
      } finally {
        inFlightRef.current = false;
      }
    }

    const intervalId = window.setInterval(() => {
      void refreshHealth();
    }, FIVE_MINUTES_MS);

    function onVisibilityChange() {
      if (document.visibilityState !== "visible") {
        return;
      }
      if (Date.now() - lastRefreshRef.current >= FIVE_MINUTES_MS) {
        void refreshHealth();
      }
    }

    document.addEventListener("visibilitychange", onVisibilityChange);
    return () => {
      window.clearInterval(intervalId);
      document.removeEventListener("visibilitychange", onVisibilityChange);
    };
  }, [router]);

  return null;
}
