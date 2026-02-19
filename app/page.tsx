import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Shield, Cloud, HardDrive, ArrowRight, Lock, Database } from "lucide-react";

export default function LandingPage() {
  return (
    <div className="flex flex-col min-h-screen bg-background text-foreground">
      <header className="px-6 py-4 flex items-center justify-between border-b sticky top-0 bg-background/80 backdrop-blur-sm z-50">
        <div className="flex items-center gap-2 font-bold text-xl">
          <Cloud className="w-6 h-6 text-foreground" />
          <span>Nebula Drive</span>
        </div>
        <div className="flex items-center gap-4">
          <Button variant="ghost" asChild>
            <Link href="/login">Log in</Link>
          </Button>
          <Button asChild>
            <Link href="/dashboard">Get Started</Link>
          </Button>
        </div>
      </header>

      <main className="flex-1">
        {/* Hero Section */}
        <section className="py-20 md:py-32 px-6 text-center max-w-4xl mx-auto">
          <h1 className="text-4xl md:text-6xl font-extrabold tracking-tight mb-6 bg-gradient-to-r from-foreground to-muted-foreground bg-clip-text text-transparent">
            RAID for your Cloud Storage
          </h1>
          <p className="text-lg md:text-xl text-muted-foreground mb-10 max-w-2xl mx-auto">
            Combine Google Drive, dropbox, and S3 into one fault-tolerant, privacy-focused storage network. 
            Zero-knowledge encryption meets redundancy.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
             <Button size="lg" className="h-12 px-8 text-lg" asChild>
               <Link href="/dashboard">
                  Start De-Clouding <ArrowRight className="ml-2 w-5 h-5" />
               </Link>
             </Button>
             <Button size="lg" variant="outline" className="h-12 w-full sm:w-auto">
                Read the Docs
             </Button>
          </div>
        </section>

        {/* Features Grid */}
        <section className="py-20 bg-muted/30 px-6">
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

      <footer className="py-8 px-6 border-t text-center text-sm text-muted-foreground">
        © 2024 Nebula Drive. Open Source and Privacy First.
      </footer>
    </div>
  );
}

function FeatureCard({ icon: Icon, title, description }: { icon: any, title: string, description: string }) {
   return (
      <div className="p-6 rounded-xl border bg-card text-card-foreground shadow-sm">
         <div className="p-3 w-fit rounded-lg bg-primary/10 text-primary mb-4">
            <Icon className="w-6 h-6" />
         </div>
         <h3 className="text-xl font-semibold mb-2">{title}</h3>
         <p className="text-muted-foreground leading-relaxed">
            {description}
         </p>
      </div>
   );
}
