import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { 
  DropdownMenu, 
  DropdownMenuContent, 
  DropdownMenuItem, 
  DropdownMenuTrigger 
} from "@/components/ui/dropdown-menu";
import { File, Download, Trash2, MoreVertical, ExternalLink } from "lucide-react";
import { HealthBadge } from "./HealthBadge";

interface FileCardProps {
  id: string;
  name: string;
  size: string;
  uploadedAt: string;
  shardsAvailable: number;
  shardsTotal: number;
  onDownload: () => void;
  onDelete: () => void;
  onDetails: () => void;
}

export function FileCard({
  name,
  size,
  uploadedAt,
  shardsAvailable,
  shardsTotal,
  onDownload,
  onDelete,
  onDetails,
}: FileCardProps) {
  return (
    <Card className="hover:bg-muted/50 transition-colors">
      <CardContent className="p-4 flex items-center gap-4">
        <div className="p-2 bg-primary/10 rounded-lg text-primary">
          <File className="w-8 h-8" />
        </div>
        
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <h3 className="font-medium truncate" title={name}>{name}</h3>
            <HealthBadge 
              shardsAvailable={shardsAvailable} 
              shardsTotal={shardsTotal} 
              minShards={4} // Default assumption for now
            />
          </div>
          <p className="text-sm text-muted-foreground flex items-center gap-2">
            <span>{size}</span>
            <span>•</span>
            <span>{uploadedAt}</span>
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" onClick={onDownload}>
            <Download className="w-4 h-4" />
          </Button>
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreVertical className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={onDetails}>
                <ExternalLink className="w-4 h-4 mr-2" />
                Details
              </DropdownMenuItem>
              <DropdownMenuItem onClick={onDelete} className="text-destructive focus:text-destructive">
                <Trash2 className="w-4 h-4 mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardContent>
    </Card>
  );
}
