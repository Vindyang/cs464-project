"use client";

import Link from "next/link";
import { ChevronLeft } from "lucide-react";

export default function BackButton() {
  return (
    <Link
      href="/"
      className="absolute top-8 left-8 z-20 flex items-center gap-1 text-sm text-zinc-400 hover:text-white transition-colors group"
    >
      <ChevronLeft className="w-4 h-4 transition-transform duration-200 group-hover:-translate-x-1" />
      Back
    </Link>
  );
}
