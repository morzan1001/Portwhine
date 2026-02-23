import {
  Radio,
  Globe,
  Radar,
  Search,
  ShieldAlert,
  GitBranch,
  Network,
  FileSearch,
  Camera,
  Lock,
  Terminal,
  Layers,
  FileCode,
  FileText,
  Database,
  HardDrive,
  Server,
  type LucideIcon,
} from 'lucide-react'

const nodeIconMap: Record<string, LucideIcon> = {
  radio: Radio,
  globe: Globe,
  radar: Radar,
  search: Search,
  'shield-alert': ShieldAlert,
  'git-branch': GitBranch,
  network: Network,
  'file-search': FileSearch,
  camera: Camera,
  lock: Lock,
  terminal: Terminal,
  layers: Layers,
  'file-code': FileCode,
  'file-text': FileText,
  database: Database,
  'hard-drive': HardDrive,
  server: Server,
}

export function getNodeIcon(iconName?: string): LucideIcon {
  if (iconName && nodeIconMap[iconName]) {
    return nodeIconMap[iconName]
  }
  return Server
}
