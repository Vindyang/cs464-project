import { Button } from "@/components/ui/button";
import { Download, Trash2, MoreHorizontal, FileText, Database } from "lucide-react";
import { DashboardCard } from "./DashboardCard";

interface File {
  id: string;
  name: string;
  size: string;
  date: string;
  providerCount: number;
  status: "synced" | "syncing" | "error";
}

interface FileTableProps {
  files: File[];
  onDownload: (id: string) => void;
  onDelete: (id: string) => void;
  onDetails: (id: string) => void;
}

export function FileTable({ files, onDownload, onDelete, onDetails }: FileTableProps) {
  return (
    <DashboardCard className="p-0 overflow-hidden bg-bg-canvas border-border-color">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="bg-bg-subtle border-b border-border-color">
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium">Name</th>
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium">Size</th>
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium">Redundancy</th>
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium">Date Modified</th>
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {files.map((file) => (
              <tr 
                key={file.id} 
                className="border-b border-border-color last:border-0 hover:bg-bg-subtle/30 transition-colors group"
              >
                <td className="p-4 align-middle">
                   <div className="flex items-center gap-3">
                      <div className="w-8 h-8 flex items-center justify-center bg-bg-subtle rounded-[2px] text-text-secondary border border-border-color">
                         <FileText className="w-4 h-4" />
                      </div>
                      <button
                        type="button"
                        className="font-medium text-left text-text-main group-hover:text-accent-primary transition-colors cursor-pointer hover:underline focus-visible:outline-none focus-visible:underline"
                        onClick={() => onDetails(file.id)}
                        aria-label={`View details for ${file.name}`}
                      >
                         {file.name}
                      </button>
                   </div>
                </td>
                <td className="p-4 align-middle font-mono text-xs text-text-secondary">
                   {file.size}
                </td>
                <td className="p-4 align-middle">
                   <div className="flex items-center gap-1.5">
                      <Database className="w-3.5 h-3.5 text-text-tertiary" />
                      <span className="text-xs text-text-secondary">{file.providerCount} Providers</span>
                   </div>
                </td>
                <td className="p-4 align-middle text-xs text-text-secondary">
                   {file.date}
                </td>
                <td className="p-4 align-middle text-right">
                   <div className="flex justify-end gap-2 opacity-100 md:opacity-0 md:group-hover:opacity-100 md:group-focus-within:opacity-100 transition-opacity">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 hover:bg-bg-subtle hover:text-accent-primary"
                        onClick={() => onDownload(file.id)}
                        aria-label={`Download ${file.name}`}
                      >
                         <Download className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 hover:bg-bg-subtle hover:text-destructive"
                        onClick={() => onDelete(file.id)}
                        aria-label={`Delete ${file.name}`}
                      >
                         <Trash2 className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 hover:bg-bg-subtle"
                        onClick={() => onDetails(file.id)}
                        aria-label={`More actions for ${file.name}`}
                      >
                         <MoreHorizontal className="w-4 h-4" />
                      </Button>
                   </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {files.length === 0 && (
         <div className="p-12 text-center text-text-secondary">
            No files found. Upload your first file to get started.
         </div>
      )}
    </DashboardCard>
  );
}
