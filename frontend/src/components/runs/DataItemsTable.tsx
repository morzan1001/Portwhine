import { useState } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Download, ChevronRight, ChevronDown, Radio } from 'lucide-react'
import { useDataItems } from '@/hooks/useDataItems'
import { timestampToDate } from '@/lib/utils'
import type { DataItemInfo } from '@/gen/portwhine/v1/operator_pb'

interface DataItemsTableProps {
  runId: string
  isStreaming?: boolean
  streamItemCount?: number
}

export function DataItemsTable({ runId, isStreaming, streamItemCount }: DataItemsTableProps) {
  const [typeFilter, setTypeFilter] = useState<string>('all')
  const [pageToken, setPageToken] = useState<string>('')
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const pageSize = 50

  const { data, isLoading } = useDataItems(
    runId,
    typeFilter === 'all' ? undefined : typeFilter,
    pageSize,
    pageToken || undefined,
  )

  const items = data?.items ?? []
  const totalCount = data?.totalCount ?? 0
  const nextPageToken = data?.nextPageToken ?? ''

  // Collect unique types from current page for the filter dropdown.
  const uniqueTypes = Array.from(new Set(items.map((item) => item.type)))

  const handleExport = () => {
    const exportData = items.map((item) => ({
      id: item.id,
      type: item.type,
      data: item.data,
      metadata: item.metadata,
      parentIds: item.parentIds,
      createdAt: item.createdAt,
    }))
    const json = JSON.stringify(exportData, null, 2)
    const blob = new Blob([json], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `data-items-${runId.substring(0, 8)}-${Date.now()}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const handleTypeChange = (value: string) => {
    setTypeFilter(value)
    setPageToken('')
    setExpandedId(null)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-medium">
            Data Items
            {totalCount > 0 && (
              <span className="text-muted-foreground font-normal ml-1.5">({totalCount})</span>
            )}
          </h3>
          {isStreaming && (
            <p className="text-xs text-muted-foreground mt-0.5 flex items-center gap-1.5">
              <Radio className="h-3 w-3 text-green-500 animate-pulse" />
              Streaming results{streamItemCount ? ` (${streamItemCount} live)` : ''}...
            </p>
          )}
        </div>
        <Button onClick={handleExport} variant="outline" size="sm" disabled={items.length === 0}>
          <Download className="mr-2 h-3.5 w-3.5" />
          Export JSON
        </Button>
      </div>

      <div className="flex gap-4">
        <Select value={typeFilter} onValueChange={handleTypeChange}>
          <SelectTrigger className="w-[200px]">
            <SelectValue placeholder="Filter by type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Types</SelectItem>
            {uniqueTypes.map((type) => (
              <SelectItem key={type} value={type}>
                {type}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {isLoading ? (
        <div className="rounded-xl border p-8 text-center">
          <p className="text-sm text-muted-foreground">Loading data items...</p>
        </div>
      ) : items.length > 0 ? (
        <>
          <div className="rounded-xl border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-8"></TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Summary</TableHead>
                  <TableHead>Source</TableHead>
                  <TableHead className="text-right">Created</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <DataItemRow
                    key={item.id}
                    item={item}
                    isExpanded={expandedId === item.id}
                    onToggle={() => setExpandedId(expandedId === item.id ? null : item.id)}
                  />
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">
              Showing {items.length} of {totalCount} items
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={!pageToken}
                onClick={() => setPageToken('')}
              >
                First Page
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={!nextPageToken}
                onClick={() => setPageToken(nextPageToken)}
              >
                Next Page
              </Button>
            </div>
          </div>
        </>
      ) : (
        <div className="text-center py-8 rounded-xl border">
          <p className="text-sm text-muted-foreground">
            {typeFilter !== 'all' ? 'No items match the selected filter' : 'No data items yet'}
          </p>
        </div>
      )}
    </div>
  )
}

function DataItemRow({
  item,
  isExpanded,
  onToggle,
}: {
  item: DataItemInfo
  isExpanded: boolean
  onToggle: () => void
}) {
  const summary = getItemSummary(item)
  const source = item.metadata?.['source'] ?? 'N/A'
  const created = timestampToDate(item.createdAt)

  return (
    <>
      <TableRow
        className="cursor-pointer hover:bg-accent/50"
        onClick={onToggle}
      >
        <TableCell className="w-8 pr-0">
          {isExpanded ? (
            <ChevronDown className="h-4 w-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          )}
        </TableCell>
        <TableCell>
          <Badge variant="secondary">{item.type}</Badge>
        </TableCell>
        <TableCell className="text-sm max-w-[400px] truncate">
          {summary}
        </TableCell>
        <TableCell className="text-sm text-muted-foreground">{source}</TableCell>
        <TableCell className="text-right text-sm text-muted-foreground">
          {created?.toLocaleString() || 'N/A'}
        </TableCell>
      </TableRow>
      {isExpanded && (
        <TableRow>
          <TableCell colSpan={5} className="bg-muted/30 p-0">
            <div className="p-4">
              <DataItemDetail item={item} />
            </div>
          </TableCell>
        </TableRow>
      )}
    </>
  )
}

function DataItemDetail({ item }: { item: DataItemInfo }) {
  const data = item.data as Record<string, unknown> | undefined

  return (
    <div className="space-y-3 text-sm">
      <div className="flex gap-8">
        <div>
          <span className="text-xs text-muted-foreground">ID</span>
          <p className="font-mono text-xs">{item.id}</p>
        </div>
        {item.parentIds.length > 0 && (
          <div>
            <span className="text-xs text-muted-foreground">Parent IDs</span>
            <p className="font-mono text-xs">{item.parentIds.map(id => id.substring(0, 12)).join(', ')}</p>
          </div>
        )}
      </div>
      {data && Object.keys(data).length > 0 && (
        <div>
          <span className="text-xs text-muted-foreground">Data</span>
          <pre className="mt-1 rounded-lg bg-muted p-3 text-xs font-mono overflow-x-auto max-h-[300px] overflow-y-auto">
            {JSON.stringify(data, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}

/** Extract a short human-readable summary from the DataItem data fields. */
function getItemSummary(item: DataItemInfo): string {
  const data = item.data as Record<string, unknown> | undefined
  if (!data) return '-'

  switch (item.type) {
    case 'domain':
      return String(data.domain ?? data.name ?? '-')
    case 'ip_address':
      return String(data.ip ?? '-')
    case 'service':
      return `${data.ip ?? ''}:${data.port ?? ''} ${data.service_name ?? ''}`.trim() || '-'
    case 'url':
      return String(data.url ?? '-')
    case 'vulnerability': {
      const severity = data.severity ? `[${data.severity}] ` : ''
      return `${severity}${data.name ?? data.template_id ?? '-'}`
    }
    case 'ssl_result':
      return String(data.host ?? data.subject ?? '-')
    case 'http_headers':
      return String(data.url ?? '-')
    case 'web_technology':
      return `${data.name ?? ''} ${data.version ?? ''}`.trim() || '-'
    case 'screenshot':
      return String(data.url ?? '-')
    case 'ssh_audit_result':
      return String(data.host ?? '-')
    case 'whois_result':
      return String(data.domain ?? data.query ?? '-')
    case 'dns_record':
      return `${data.type ?? ''} ${data.name ?? ''} → ${data.value ?? ''}`.trim() || '-'
    case 'report':
      return `${data.title ?? 'Report'} (${data.total_items ?? 0} items)`
    case 'webhook_delivery':
      return `${data.method ?? 'POST'} ${data.url ?? ''} → ${data.status_code ?? '?'}`
    case 'email_delivery':
      return `${data.status ?? 'sent'} to ${Array.isArray(data.recipients) ? (data.recipients as string[]).join(', ') : '?'}`
    default: {
      // Attempt to find a meaningful field.
      for (const key of ['name', 'url', 'host', 'ip', 'domain', 'title']) {
        if (data[key]) return String(data[key])
      }
      return Object.keys(data).slice(0, 3).join(', ') || '-'
    }
  }
}
