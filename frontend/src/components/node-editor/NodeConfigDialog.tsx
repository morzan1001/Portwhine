import React, { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from '@/components/ui/form'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { JsonSchemaForm } from './JsonSchemaForm'
import { LogViewer } from '@/components/runs/LogViewer'
import { useStreamNodeLogs } from '@/hooks/useStreaming'
import { getNodeIcon } from '@/lib/node-editor/icons'
import { Trash2 } from 'lucide-react'
import type { ReactFlowNodeData } from '@/lib/node-editor/conversions'
import { PipelineNodeType } from '@/gen/portwhine/v1/pipeline_pb'

const nodeConfigSchema = z.object({
  label: z.string().min(1, 'Label is required'),
  replicas: z.number().int().min(1).max(10),
  inputFilterType: z.string().optional(),
  inputFilterCondition: z.string().optional(),
  retryMaxRetries: z.number().int().min(0).max(10).optional(),
  retryInitialBackoff: z.number().int().min(1).optional(),
  retryMaxBackoff: z.number().int().min(1).optional(),
})

type NodeConfigForm = z.infer<typeof nodeConfigSchema>

interface NodeConfigDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  nodeData: ReactFlowNodeData | null
  onSave: (data: Partial<ReactFlowNodeData>) => void
  onDelete?: (nodeId: string) => void
  runId?: string
}

export function NodeConfigDialog({
  open,
  onOpenChange,
  nodeData,
  onSave,
  onDelete,
  runId,
}: NodeConfigDialogProps) {
  const [configValues, setConfigValues] = useState<Record<string, unknown>>({})
  const [rawConfig, setRawConfig] = useState('{}')
  const hasSchema = !!nodeData?.configSchema
  const hasLogs = !!runId && !!nodeData

  const { logs, isStreaming } = useStreamNodeLogs(
    runId || '',
    nodeData?.id || '',
    hasLogs && open
  )

  const form = useForm<NodeConfigForm>({
    resolver: zodResolver(nodeConfigSchema),
    defaultValues: {
      label: '',
      replicas: 1,
      inputFilterType: '',
      inputFilterCondition: '',
      retryMaxRetries: 3,
      retryInitialBackoff: 1,
      retryMaxBackoff: 60,
    },
  })

  const [prevNodeId, setPrevNodeId] = useState<string | null>(null)
  if (nodeData && nodeData.id !== prevNodeId) {
    setPrevNodeId(nodeData.id)
    form.reset({
      label: nodeData.label || '',
      replicas: nodeData.replicas || 1,
      inputFilterType: nodeData.inputFilter?.type || '',
      inputFilterCondition: nodeData.inputFilter?.condition || '',
      retryMaxRetries: nodeData.retryPolicy?.maxRetries || 3,
      retryInitialBackoff: nodeData.retryPolicy?.initialBackoffSeconds || 1,
      retryMaxBackoff: nodeData.retryPolicy?.maxBackoffSeconds || 60,
    })
    const cfg = (nodeData.config || {}) as Record<string, unknown>
    setConfigValues(cfg)
    setRawConfig(JSON.stringify(cfg, null, 2))
  }

  const handleSubmit = (data: NodeConfigForm) => {
    let config: Record<string, unknown>
    if (hasSchema) {
      config = configValues
    } else {
      try {
        config = JSON.parse(rawConfig)
      } catch {
        return
      }
    }

    const updates: Partial<ReactFlowNodeData> = {
      label: data.label,
      replicas: data.replicas,
      config,
    }

    if (data.inputFilterType || data.inputFilterCondition) {
      updates.inputFilter = {
        type: data.inputFilterType,
        condition: data.inputFilterCondition,
      }
    }

    if (data.retryMaxRetries !== undefined) {
      updates.retryPolicy = {
        maxRetries: data.retryMaxRetries,
        initialBackoffSeconds: data.retryInitialBackoff || 1,
        maxBackoffSeconds: data.retryMaxBackoff || 60,
      }
    }

    onSave(updates)
    onOpenChange(false)
  }

  if (!nodeData) return null

  const isWorkerNode = nodeData.type === PipelineNodeType.WORKER
  const tabCount = 2 + (isWorkerNode ? 1 : 0) + (hasLogs ? 1 : 0)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={`${hasLogs ? 'max-w-3xl' : 'max-w-2xl'} max-h-[90vh] overflow-y-auto`}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {nodeData.color && (
              <div
                className="flex h-6 w-6 items-center justify-center rounded"
                style={{ backgroundColor: `${nodeData.color}20`, color: nodeData.color }}
              >
                {React.createElement(getNodeIcon(nodeData.icon), { className: "h-3.5 w-3.5" })}
              </div>
            )}
            Configure {nodeData.label}
            {nodeData.catalogId && (
              <Badge variant="outline" className="text-xs ml-1">
                {nodeData.catalogId}
              </Badge>
            )}
          </DialogTitle>
          <DialogDescription>
            {nodeData.image && (
              <span className="font-mono text-xs">{nodeData.image}</span>
            )}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
            <Tabs defaultValue="basic" className="w-full">
              <TabsList className={`grid w-full`} style={{ gridTemplateColumns: `repeat(${tabCount}, 1fr)` }}>
                <TabsTrigger value="basic">Basic</TabsTrigger>
                <TabsTrigger value="config">Config</TabsTrigger>
                {isWorkerNode && <TabsTrigger value="advanced">Advanced</TabsTrigger>}
                {hasLogs && <TabsTrigger value="logs">Logs</TabsTrigger>}
              </TabsList>

              {/* ── Basic Tab ──────────────────────────── */}
              <TabsContent value="basic" className="space-y-4 mt-4">
                <FormField
                  control={form.control}
                  name="label"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Label</FormLabel>
                      <FormControl>
                        <Input placeholder="Node label" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                {isWorkerNode && (
                  <FormField
                    control={form.control}
                    name="replicas"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Replicas</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            min={1}
                            max={10}
                            {...field}
                            onChange={(e) => field.onChange(parseInt(e.target.value))}
                          />
                        </FormControl>
                        <FormDescription>
                          Number of parallel workers (1-10)
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}

                {/* Read-only info */}
                {nodeData.acceptedInputTypes && nodeData.acceptedInputTypes.length > 0 && (
                  <div>
                    <p className="text-sm font-medium mb-1">Accepted Inputs</p>
                    <div className="flex gap-1 flex-wrap">
                      {nodeData.acceptedInputTypes.map((t) => (
                        <Badge key={t} variant="secondary" className="text-xs font-mono">
                          {t}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}

                {nodeData.outputTypes && nodeData.outputTypes.length > 0 && (
                  <div>
                    <p className="text-sm font-medium mb-1">Outputs</p>
                    <div className="flex gap-1 flex-wrap">
                      {nodeData.outputTypes.map((t) => (
                        <Badge key={t} variant="outline" className="text-xs font-mono">
                          {t}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}
              </TabsContent>

              {/* ── Config Tab ─────────────────────────── */}
              <TabsContent value="config" className="space-y-4 mt-4">
                {hasSchema ? (
                  <JsonSchemaForm
                    schema={nodeData.configSchema!}
                    values={configValues}
                    onChange={setConfigValues}
                  />
                ) : (
                  <div className="space-y-2">
                    <p className="text-sm font-medium">Configuration (JSON)</p>
                    <Textarea
                      value={rawConfig}
                      onChange={(e) => setRawConfig(e.target.value)}
                      placeholder='{"key": "value"}'
                      className="font-mono text-sm min-h-[200px]"
                    />
                    <p className="text-xs text-muted-foreground">
                      Worker-specific configuration as JSON
                    </p>
                  </div>
                )}
              </TabsContent>

              {/* ── Advanced Tab (Workers only) ────────── */}
              {isWorkerNode && (
                <TabsContent value="advanced" className="space-y-4 mt-4">
                  <div className="space-y-4">
                    <h3 className="text-sm font-medium">Input Filter</h3>
                    <FormField
                      control={form.control}
                      name="inputFilterType"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Data Type Filter</FormLabel>
                          <FormControl>
                            <Input placeholder="e.g., domain, ip, url" {...field} />
                          </FormControl>
                          <FormDescription>Only process items of this type</FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="inputFilterCondition"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Condition</FormLabel>
                          <FormControl>
                            <Textarea placeholder="e.g., data.port == 443" {...field} />
                          </FormControl>
                          <FormDescription>Optional filter expression</FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>

                  <div className="space-y-4 mt-6">
                    <h3 className="text-sm font-medium">Retry Policy</h3>
                    <FormField
                      control={form.control}
                      name="retryMaxRetries"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Max Retries</FormLabel>
                          <FormControl>
                            <Input
                              type="number"
                              min={0}
                              max={10}
                              {...field}
                              onChange={(e) => field.onChange(parseInt(e.target.value))}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="retryInitialBackoff"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Initial Backoff (seconds)</FormLabel>
                          <FormControl>
                            <Input
                              type="number"
                              min={1}
                              {...field}
                              onChange={(e) => field.onChange(parseInt(e.target.value))}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="retryMaxBackoff"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Max Backoff (seconds)</FormLabel>
                          <FormControl>
                            <Input
                              type="number"
                              min={1}
                              {...field}
                              onChange={(e) => field.onChange(parseInt(e.target.value))}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                </TabsContent>
              )}

              {/* ── Logs Tab ──────────────────────────── */}
              {hasLogs && (
                <TabsContent value="logs" className="mt-4">
                  <div className="h-[350px]">
                    <LogViewer
                      logs={logs}
                      isStreaming={isStreaming}
                      title={`${nodeData.label} Logs`}
                    />
                  </div>
                </TabsContent>
              )}
            </Tabs>

            <div className="flex justify-between">
              {onDelete && nodeData ? (
                <Button
                  type="button"
                  variant="ghost"
                  className="text-destructive hover:text-destructive"
                  onClick={() => onDelete(nodeData.id)}
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  Delete Node
                </Button>
              ) : (
                <div />
              )}
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => onOpenChange(false)}
                >
                  Cancel
                </Button>
                <Button type="submit">Save Changes</Button>
              </div>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
