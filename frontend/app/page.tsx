import Link from "next/link";
import { Shield, Lock, Database, ArrowRight, Terminal, Server, Layers, Cpu, Network } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function LandingPage() {
  return (
    <div className="flex min-h-screen flex-col bg-white">
      {/* Header */}
      <header className="sticky top-0 z-50 border-b bg-white/80 backdrop-blur-sm">
        <div className="container mx-auto flex h-16 items-center justify-between px-6">
          <div className="flex items-center gap-2">
            <div className="flex h-6 w-6 items-center justify-center border-2 border-black">
              <div className="h-2 w-2 bg-black" />
            </div>
            <span className="font-mono text-sm font-bold tracking-wider">
              NEBULA_DRIVE
            </span>
          </div>

          <div className="flex items-center gap-6">
            <Link
              href="/login"
              className="font-mono text-sm text-neutral-600 transition-colors hover:text-black"
            >
              Log in
            </Link>
            <Button asChild size="sm" className="font-mono text-sm">
              <Link href="/dashboard">Get Started</Link>
            </Button>
          </div>
        </div>
      </header>

      <main className="flex-1">
        {/* Hero Section */}
        <section className="container mx-auto px-6 py-24 text-center">
          <div className="mb-8 inline-flex items-center gap-2 border border-black px-3 py-1">
            <div className="h-1.5 w-1.5 bg-black" />
            <span className="font-mono text-xs uppercase tracking-wider">
              v0.1-alpha
            </span>
          </div>

          <h1 className="mb-6 text-6xl font-bold leading-tight tracking-tight md:text-7xl lg:text-8xl">
            Distributed
            <br />
            Cloud Storage
          </h1>

          <p className="mx-auto mb-12 max-w-2xl font-mono text-sm leading-relaxed text-neutral-600">
            RAID architecture for the cloud. Encrypt locally, distribute globally,
            survive provider failures. Zero-knowledge. Fault-tolerant. Developer-first.
          </p>

          <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Button size="lg" className="font-mono" asChild>
              <Link href="/dashboard">
                <Terminal className="mr-2 h-4 w-4" />
                Start System
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>

            <Button size="lg" variant="outline" className="font-mono" asChild>
              <Link href="#specs">
                <Server className="mr-2 h-4 w-4" />
                Read Specs
              </Link>
            </Button>
          </div>

          {/* Tech specs preview */}
          <div className="mt-20 grid gap-4 sm:grid-cols-3">
            <TechStat label="Reed-Solomon" value="(6,4)" />
            <TechStat label="Encryption" value="AES-256" />
            <TechStat label="Min Shards" value="66%" />
          </div>
        </section>

        {/* Features Section */}
        <section id="specs" className="border-t bg-neutral-50 py-24">
          <div className="container mx-auto px-6">
            <div className="mb-16 text-center">
              <h2 className="mb-4 font-mono text-xs uppercase tracking-wider text-neutral-600">
                System Capabilities
              </h2>
              <p className="text-3xl font-bold md:text-4xl">
                Built for developers who need
                <br />
                transparency and control
              </p>
            </div>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              <FeatureCard
                icon={Lock}
                title="Client-Side Encryption"
                description="Files encrypted in-browser with AES-256-GCM before upload. Your keys, your data. Providers see encrypted shards only."
              />
              <FeatureCard
                icon={Layers}
                title="Reed-Solomon Sharding"
                description="(6,4) erasure coding splits files into shards. Lose 2 providers, still recover 100% of data. RAID for the cloud."
              />
              <FeatureCard
                icon={Shield}
                title="Provider Agnostic"
                description="Distribute across Google Drive, AWS S3, Dropbox. No single point of failure. No vendor lock-in."
              />
              <FeatureCard
                icon={Network}
                title="Real-Time Monitoring"
                description="Live shard health tracking. Provider status dashboards. Upload progress per-shard. Full transparency."
              />
              <FeatureCard
                icon={Cpu}
                title="Web Crypto API"
                description="Browser-native encryption. No server-side keys. Crypto operations in Web Workers for performance."
              />
              <FeatureCard
                icon={Database}
                title="Fault Tolerance"
                description="System survives provider outages. Files downloadable with any 4 of 6 shards. Redundancy by design."
              />
            </div>
          </div>
        </section>

        {/* Architecture */}
        <section className="border-t py-24">
          <div className="container mx-auto px-6">
            <div className="mx-auto max-w-4xl">
              <h2 className="mb-12 text-center font-mono text-xs uppercase tracking-wider text-neutral-600">
                How It Works
              </h2>

              <div className="space-y-6">
                <ProcessStep
                  number="01"
                  title="Encrypt Locally"
                  description="Files encrypted in your browser with AES-256-GCM. Encryption key never leaves your device."
                />
                <ProcessStep
                  number="02"
                  title="Create Shards"
                  description="Reed-Solomon encoding splits encrypted file into 6 shards (4 data + 2 parity). Need any 4 to reconstruct."
                />
                <ProcessStep
                  number="03"
                  title="Distribute"
                  description="Shards uploaded to multiple providers in parallel. Each provider stores encrypted shards—unreadable and incomplete."
                />
                <ProcessStep
                  number="04"
                  title="Monitor & Recover"
                  description="Real-time health monitoring. Download any 4 shards, reconstruct file, decrypt locally. Providers can fail—your data survives."
                />
              </div>
            </div>
          </div>
        </section>

        {/* CTA */}
        <section className="border-t bg-neutral-50 py-24">
          <div className="container mx-auto px-6 text-center">
            <h2 className="mb-6 text-4xl font-bold">
              Ready to distribute your data?
            </h2>
            <p className="mx-auto mb-8 max-w-xl font-mono text-sm text-neutral-600">
              Connect your cloud providers, upload your first file, and experience
              fault-tolerant storage.
            </p>
            <Button size="lg" className="font-mono" asChild>
              <Link href="/dashboard">
                <Terminal className="mr-2 h-4 w-4" />
                Initialize System
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t py-8">
        <div className="container mx-auto px-6 text-center">
          <p className="font-mono text-xs text-neutral-500">
            © 2024 NEBULA_DRIVE • Open Source • Privacy First
          </p>
        </div>
      </footer>
    </div>
  );
}

function TechStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="border bg-white p-6">
      <div className="mb-2 font-mono text-xs uppercase tracking-wider text-neutral-500">
        {label}
      </div>
      <div className="font-mono text-3xl font-bold">{value}</div>
    </div>
  );
}

function FeatureCard({
  icon: Icon,
  title,
  description,
}: {
  icon: any;
  title: string;
  description: string;
}) {
  return (
    <div className="border bg-white p-6 transition-all hover:shadow-lg">
      <div className="mb-4 inline-flex h-12 w-12 items-center justify-center border bg-neutral-50">
        <Icon className="h-6 w-6" />
      </div>
      <h3 className="mb-3 font-mono text-sm font-bold uppercase tracking-wider">
        {title}
      </h3>
      <p className="font-mono text-xs leading-relaxed text-neutral-600">
        {description}
      </p>
    </div>
  );
}

function ProcessStep({
  number,
  title,
  description,
}: {
  number: string;
  title: string;
  description: string;
}) {
  return (
    <div className="flex gap-6 border bg-white p-6">
      <div className="flex h-12 w-12 flex-shrink-0 items-center justify-center border-2 border-black bg-black font-mono text-lg font-bold text-white">
        {number}
      </div>
      <div>
        <h3 className="mb-2 font-mono text-lg font-bold">{title}</h3>
        <p className="font-mono text-sm leading-relaxed text-neutral-600">
          {description}
        </p>
      </div>
    </div>
  );
}
