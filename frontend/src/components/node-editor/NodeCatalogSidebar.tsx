import { useMemo, useState } from 'react'
import { useNodeCatalog } from '@/hooks/useNodeCatalog'
import { CatalogEntryCard } from './CatalogEntryCard'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Search, Loader2 } from 'lucide-react'

const CATEGORY_ORDER: Record<string, number> = {
  monitoring: 0,
  input: 1,
  scanning: 2,
  enumeration: 3,
  analysis: 4,
  reporting: 5,
  output: 6,
}

const CATEGORY_LABELS: Record<string, string> = {
  monitoring: 'Monitoring',
  input: 'Input',
  scanning: 'Scanning',
  enumeration: 'Enumeration',
  analysis: 'Analysis',
  reporting: 'Reporting',
  output: 'Output',
}

export function NodeCatalogSidebar() {
  const { data: catalog, isLoading } = useNodeCatalog()
  const [search, setSearch] = useState('')

  const grouped = useMemo(() => {
    if (!catalog) return new Map<string, typeof catalog>()

    const filtered = search
      ? catalog.filter(
          (e) =>
            e.displayName.toLowerCase().includes(search.toLowerCase()) ||
            e.description.toLowerCase().includes(search.toLowerCase()) ||
            e.category.toLowerCase().includes(search.toLowerCase())
        )
      : catalog

    const groups = new Map<string, typeof catalog>()
    for (const entry of filtered) {
      const key = entry.category
      if (!groups.has(key)) groups.set(key, [])
      groups.get(key)!.push(entry)
    }

    // Sort by category order
    return new Map(
      [...groups.entries()].sort(
        ([a], [b]) => (CATEGORY_ORDER[a] ?? 99) - (CATEGORY_ORDER[b] ?? 99)
      )
    )
  }, [catalog, search])

  return (
    <div className="w-[280px] border-r bg-card flex flex-col shrink-0">
      <div className="p-3 border-b">
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search nodes..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-8 h-9"
          />
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-3 space-y-4">
          {isLoading && (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          )}

          {!isLoading && grouped.size === 0 && (
            <p className="text-sm text-muted-foreground text-center py-8">
              No nodes found
            </p>
          )}

          {[...grouped.entries()].map(([category, entries]) => (
            <div key={category}>
              <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                {CATEGORY_LABELS[category] || category}
              </h3>
              <div className="space-y-1.5">
                {entries?.map((entry) => (
                  <CatalogEntryCard key={entry.id} entry={entry} />
                ))}
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  )
}
