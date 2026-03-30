import { StorageOverview } from "@/components/dashboard/StorageOverview";
import { ProviderMatrix } from "@/components/dashboard/ProviderMatrix";
import { SystemHealth } from "@/components/dashboard/SystemHealth";
import { ActivityFeed } from "@/components/dashboard/ActivityFeed";
import { TechSpecs } from "@/components/dashboard/TechSpecs";
import { getDashboardData } from "./componentsAction/actions";

export default async function DashboardPage() {
  const { files, providers } = await getDashboardData();

  return (
    <div className="grid grid-cols-1 lg:grid-cols-[280px_1fr_320px] gap-6 max-w-[1600px] mx-auto">
      {/* Column 1 */}
      <div className="flex flex-col gap-6">
        <StorageOverview providers={providers} fileCount={files.length} />
        <ActivityFeed />
      </div>

      {/* Column 2 */}
      <div>
        <ProviderMatrix providers={providers} />
      </div>

      {/* Column 3 */}
      <div className="flex flex-col gap-6">
        <SystemHealth />
        <TechSpecs />
      </div>
    </div>
  );
}
