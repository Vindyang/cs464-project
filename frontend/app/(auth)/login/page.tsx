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

export default function LoginPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsLoading(true);

    // Simulate login delay
    setTimeout(() => {
      setIsLoading(false);
      toast.success("Logged in successfully");
      router.push("/dashboard");
    }, 1000);
  }

  return (
    <DashboardCard className="w-full">
      <div className="flex flex-col space-y-1.5 p-6 pb-2">
        <h3 className="font-semibold tracking-tight text-xl text-text-main">Welcome back</h3>
        <p className="text-sm text-text-secondary">
          Enter your credentials to access your vault
        </p>
      </div>
      <div className="p-6 pt-0">
        <form onSubmit={onSubmit}>
          <div className="grid gap-4">
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
              <div className="flex items-center">
                <Label htmlFor="password">Password</Label>
                <Link
                  href="/forgot-password"
                  className="ml-auto inline-block text-xs font-medium text-accent-primary hover:underline hover:text-accent-primary-hover"
                >
                  Forgot your password?
                </Link>
              </div>
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
              Sign In with Email
            </Button>
            <Button variant="outline" type="button" disabled={isLoading} className="border-border-color text-text-main hover:bg-bg-subtle rounded-[2px]">
               Continue with Google
            </Button>
          </div>
        </form>
        <div className="mt-6 text-center text-sm text-text-secondary">
          Don&apos;t have an account?{" "}
          <Link href="/signup" className="font-medium text-accent-primary hover:underline hover:text-accent-primary-hover">
            Sign up
          </Link>
        </div>
      </div>
    </DashboardCard>
  );
}
