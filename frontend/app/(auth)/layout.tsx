import Iridescence from "@/components/Iridescence";
import BackButton from "@/components/BackButton";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="h-screen overflow-hidden w-full lg:grid lg:grid-cols-2">
      {/* Left Panel: Iridescence background */}
      <div className="relative hidden lg:flex overflow-hidden" style={{ filter: "grayscale(1)" }}>
        <Iridescence
          color={[0.15, 0.15, 0.15]}
          speed={0.2}
          amplitude={0.1}
          mouseReact={false}
        />

        {/* Back button — top left */}
        <BackButton />
      </div>

      {/* Right Panel: Clean, White, Form-focused */}
      <div className="flex flex-col items-center justify-center p-6 sm:p-12 relative bg-white overflow-hidden">
        {children}
      </div>
    </div>
  );
}
