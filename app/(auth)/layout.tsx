import { Cloud } from "lucide-react";
import Link from "next/link";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-background p-4">
      <div className="w-full max-w-sm space-y-6">
        <Link href="/" className="flex items-center justify-center gap-2 font-bold text-xl mb-8">
          <Cloud className="w-8 h-8 text-foreground" />
          <span>Nebula Drive</span>
        </Link>
        {children}
      </div>
    </div>
  );
}
