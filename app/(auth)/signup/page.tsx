"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DashboardCard } from "@/components/dashboard/DashboardCard";
import Link from "next/link";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";

export default function SignupPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsLoading(true);

    // Simulate signup delay
    setTimeout(() => {
      setIsLoading(false);
      toast.success("Account created successfully");
      router.push("/onboarding");
    }, 1000);
  }

  return (
    <DashboardCard className="w-full">
      <div className="flex flex-col space-y-1.5 p-6 pb-2">
        <h3 className="font-semibold tracking-tight text-xl text-text-main">Create an account</h3>
        <p className="text-sm text-text-secondary">
          Enter your email below to create your account
        </p>
      </div>
      <div className="p-6 pt-0">
        <form onSubmit={onSubmit}>
          <div className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Full Name</Label>
              <Input
                id="name"
                placeholder="John Doe"
                type="text"
                autoCapitalize="words"
                autoComplete="name"
                autoCorrect="off"
                disabled={isLoading}
                className="bg-bg-subtle border-border-color focus-visible:ring-accent-primary"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                placeholder="name@example.com"
                type="email"
                autoCapitalize="none"
                autoComplete="email"
                autoCorrect="off"
                disabled={isLoading}
                className="bg-bg-subtle border-border-color focus-visible:ring-accent-primary"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">Password</Label>
              <Input 
                  id="password" 
                  type="password" 
                  disabled={isLoading} 
                  className="bg-bg-subtle border-border-color focus-visible:ring-accent-primary"
              />
            </div>
            <Button disabled={isLoading} className="bg-accent-primary hover:bg-accent-primary-hover text-white rounded-[2px] mt-2">
              {isLoading && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              Create Account
            </Button>
            <Button variant="outline" type="button" disabled={isLoading} className="border-border-color text-text-main hover:bg-bg-subtle rounded-[2px]">
               Sign up with Google
            </Button>
          </div>
        </form>
        <div className="mt-6 text-center text-sm text-text-secondary">
          Already have an account?{" "}
          <Link href="/login" className="font-medium text-accent-primary hover:underline hover:text-accent-primary-hover">
            Log in
          </Link>
        </div>
      </div>
    </DashboardCard>
  );
}
