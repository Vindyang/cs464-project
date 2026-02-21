import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Shield, Cloud, HardDrive, ArrowRight, Lock, Database } from "lucide-react";
import { GridBackground } from "@/components/ui/grid-background";
import { DashboardCard } from "@/components/dashboard/DashboardCard";

export default function LandingPage() {
  return (
    <div className="flex flex-col min-h-screen bg-bg-canvas text-text-main font-sans relative isolate">
      <GridBackground />
      
      <header className="px-6 h-16 flex items-center justify-between border-b border-border-color bg-bg-canvas/80 backdrop-blur-sm z-50 sticky top-0">
        <div className="flex items-center gap-3 font-semibold text-lg">
          <div className="w-6 h-6 border-[1.5px] border-text-main relative flex items-center justify-center">
             <div className="w-2 h-2 bg-accent-primary" />
          </div>
          <span className="tracking-tight">ZERO-STORE</span>
        </div>
        <div className="flex items-center gap-6 text-sm font-medium">
          <Link href="/login" className="text-text-secondary hover:text-text-main transition-colors">
            Log in
          </Link>
          <Link 
            href="/dashboard"
            className="inline-flex items-center justify-center h-9 px-5 bg-accent-primary text-white rounded-[2px] transition-all hover:bg-accent-primary-hover"
          >
            Get Started
          </Link>
        </div>
      </header>

      <main className="flex-1">
        {/* Hero Section */}
        <section className="py-24 px-6 text-center max-w-4xl mx-auto">
          <div className="inline-block mb-4 px-3 py-1 bg-bg-subtle border border-border-color rounded-[2px] text-[11px] font-mono uppercase tracking-wider text-text-secondary">
             Pre-Alpha Release v0.1
          </div>
          <h1 className="text-5xl md:text-6xl font-semibold tracking-[-0.03em] mb-6 text-text-main">
            RAID for your Cloud Storage
          </h1>
          <p className="text-lg text-text-secondary mb-10 max-w-2xl mx-auto leading-relaxed">
            Combine Google Drive, dropbox, and S3 into one fault-tolerant, privacy-focused storage network. 
            Zero-knowledge encryption meets redundancy.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
             <Link 
                href="/dashboard"
                className="inline-flex items-center justify-center h-12 px-8 bg-accent-primary text-white text-base font-medium rounded-[2px] transition-all hover:bg-accent-primary-hover"
             >
                Start De-Clouding <ArrowRight className="ml-2 w-4 h-4" />
             </Link>
             <Button variant="outline" className="h-12 px-8 rounded-[2px] border-border-color text-text-main hover:bg-bg-subtle" asChild>
                <Link href="#">Read the Specs</Link>
             </Button>
          </div>
        </section>

        {/* Features Grid */}
        <section className="py-24 px-6 border-t border-border-color bg-bg-subtle/30">
           <div className="max-w-6xl mx-auto grid md:grid-cols-3 gap-8">
              <FeatureCard 
                 icon={Lock}
                 title="Zero-Knowledge Encryption"
                 description="Files are encrypted in your browser before they ever leave your device. Not even we can see your data."
              />
              <FeatureCard 
                 icon={Database}
                 title="Data Redundancy"
                 description="We split your files into shards using Reed-Solomon coding. You can lose a provider and still recover your data."
              />
              <FeatureCard 
                 icon={Shield}
                 title="Provider Agnostic"
                 description="Don't rely on a single giant. Distribute your risk across Google, AWS, and more."
              />
           </div>
        </section>
      </main>

      <footer className="py-8 px-6 border-t border-border-color text-center text-xs font-mono text-text-tertiary">
        © 2024 Nebula Drive. Open Source and Privacy First.
      </footer>
    </div>
  );
}

function FeatureCard({ icon: Icon, title, description }: { icon: any, title: string, description: string }) {
   return (
      <DashboardCard className="h-full bg-bg-canvas">
         <div className="p-3 w-fit rounded-[2px] bg-bg-subtle border border-border-color text-text-main mb-4">
            <Icon className="w-5 h-5" />
         </div>
         <h3 className="text-lg font-semibold mb-2 tracking-tight">{title}</h3>
         <p className="text-text-secondary leading-relaxed text-sm">
            {description}
         </p>
      </DashboardCard>
   );
}
