"use client";

import Link from "next/link";
import { ChevronLeft } from "lucide-react";
import { useRef } from "react";
import gsap from "gsap";

export default function BackButton() {
  const iconRef = useRef<SVGSVGElement>(null);

  function onEnter() {
    gsap.to(iconRef.current, { x: -4, duration: 0.2, ease: "power2.out" });
  }

  function onLeave() {
    gsap.to(iconRef.current, { x: 0, duration: 0.2, ease: "power2.out" });
  }

  return (
    <Link
      href="/"
      className="absolute top-8 left-8 z-20 flex items-center gap-1 text-sm text-zinc-400 hover:text-white transition-colors"
      onMouseEnter={onEnter}
      onMouseLeave={onLeave}
    >
      <ChevronLeft ref={iconRef} className="w-4 h-4" />
      Back
    </Link>
  );
}
