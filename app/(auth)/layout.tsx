import Link from "next/link";
import { GridBackground } from "@/components/ui/grid-background";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-bg-canvas p-4 relative isolate">
      <GridBackground />
      <div className="w-full max-w-sm space-y-8 relative z-10">
        <Link href="/" className="flex items-center justify-center gap-3 font-semibold text-xl mb-8 text-text-main">
          <div className="w-8 h-8 border-[2px] border-text-main relative flex items-center justify-center bg-bg-canvas">
             <div className="w-3 h-3 bg-accent-primary" />
          </div>
          <span className="tracking-tight">Nebula Drive </span>
        </Link>
        {children}
      </div>
    </div>
  );
}
