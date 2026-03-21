"use client"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { signIn } from "@/lib/auth-client"
import { useState } from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { Loader2 } from "lucide-react"
import Link from "next/link"

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setIsLoading(true)

    try {
      const result = await signIn.email({
        email,
        password,
      })

      if (result.error) {
        toast.error(result.error.message || "Invalid email or password")
        setIsLoading(false)
        return
      }

      toast.success("Logged in successfully")
      router.push("/dashboard")
    } catch {
      toast.error("Something went wrong. Please try again.")
      setIsLoading(false)
    }
  }

  return (
    <div className={cn("flex flex-col gap-8 w-full max-w-[400px] mx-auto animate-in fade-in slide-in-from-bottom-4 duration-700", className)} {...props}>
      <div className="flex flex-col space-y-2 text-center sm:text-left">
        <h1 className="text-3xl font-semibold tracking-tight text-neutral-950">
          Welcome back
        </h1>
        <p className="text-sm text-neutral-500">
          Enter your credentials to access your account
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-6">
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email" className="text-neutral-700 font-medium">Email Address</Label>
            <Input
              id="email"
              type="email"
              placeholder="name@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isLoading}
              required
              className="h-11 px-4 border-neutral-200 bg-white focus-visible:ring-1 focus-visible:ring-black focus-visible:border-black rounded-md transition-all placeholder:text-neutral-400"
            />
          </div>
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="password" className="text-neutral-700 font-medium">Password</Label>
              <Link
                href="/forgot-password"
                className="text-sm font-medium text-neutral-500 hover:text-black transition-colors"
                tabIndex={-1}
              >
                Forgot password?
              </Link>
            </div>
            <Input
              id="password"
              type="password"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isLoading}
              required
              className="h-11 px-4 border-neutral-200 bg-white focus-visible:ring-1 focus-visible:ring-black focus-visible:border-black rounded-md transition-all placeholder:text-neutral-400"
            />
          </div>
        </div>

        <Button 
          type="submit" 
          disabled={isLoading} 
          className="w-full h-11 bg-black text-white hover:bg-neutral-800 focus-visible:ring-2 focus-visible:ring-neutral-400 focus-visible:ring-offset-2 active:scale-[0.98] transition-all rounded-md font-medium"
        >
          {isLoading ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            "Sign In"
          )}
        </Button>
      </form>

      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <span className="w-full border-t border-neutral-200" />
        </div>
        <div className="relative flex justify-center text-xs uppercase">
          <span className="bg-white px-3 text-neutral-500 font-medium">
            Or continue with
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4">
        <Button 
          variant="outline" 
          type="button" 
          disabled={isLoading}
          className="h-11 bg-white border-neutral-200 hover:bg-neutral-50 hover:border-neutral-300 text-neutral-700 font-medium rounded-md transition-all"
        >
          <svg className="mr-2 h-4 w-4" aria-hidden="true" focusable="false" data-prefix="fab" data-icon="google" role="img" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 488 512">
            <path fill="currentColor" d="M488 261.8C488 403.3 391.1 504 248 504 110.8 504 0 393.2 0 256S110.8 8 248 8c66.8 0 123 24.5 166.3 64.9l-67.5 64.9C258.5 52.6 94.3 116.6 94.3 256c0 86.5 69.1 156.6 153.7 156.6 98.2 0 135-70.4 140.8-106.9H248v-85.3h236.1c2.3 12.7 3.9 24.9 3.9 41.4z"></path>
          </svg>
          Google
        </Button>
      </div>

      <p className="text-center text-sm text-neutral-500">
        Don&apos;t have an account?{" "}
        <Link href="/signup" className="font-semibold text-black hover:underline underline-offset-4">
          Create an account
        </Link>
      </p>
    </div>
  )
}
