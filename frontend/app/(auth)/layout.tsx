import Link from "next/link";
import { Mountain } from "lucide-react"; // Using Mountain as a placeholder luxury icon

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen w-full lg:grid lg:grid-cols-2">
      {/* Left Panel: Dark, Branded, Premium aesthetic */}
      <div className="relative hidden flex-col bg-zinc-950 p-10 text-white lg:flex justify-between overflow-hidden">
        {/* Abstract background */}
        <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)] opacity-30" />
        
        <div className="relative z-20 flex items-center gap-3 text-2xl font-bold tracking-tight">
          <div className="flex items-center justify-center p-1 bg-white rounded-md">
            <Mountain className="w-6 h-6 text-zinc-950" strokeWidth={2.5} />
          </div>
          Nebula Drive
        </div>
        
        <div className="relative z-20 mt-auto">
          <blockquote className="space-y-4">
            <p className="text-2xl font-medium tracking-tight leading-snug">
              &ldquo;Building the future of academic orchestration. Fast, reliable, and exceptionally intuitive.&rdquo;
            </p>
            <footer className="text-sm font-light text-zinc-400">
              The CS464 Orchestration Team
            </footer>
          </blockquote>
        </div>
      </div>

      {/* Right Panel: Clean, White, Form-focused */}
      <div className="flex flex-col items-center justify-center p-6 sm:p-12 relative bg-white overflow-hidden">
        {/* Mobile Logo Only */}
        <Link href="/" className="lg:hidden flex items-center justify-center gap-2 font-bold text-xl mb-12 text-black">
          <Mountain className="w-5 h-5" strokeWidth={2.5} />
          Nebula Drive
        </Link>
        {children}
      </div>
    </div>
  );
}
