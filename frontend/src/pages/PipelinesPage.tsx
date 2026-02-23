import { useState } from 'react'
import { usePipelines, useDeletePipeline } from '@/hooks/usePipelines'
import { PipelineTable } from '@/components/pipelines/PipelineTable'
import { CreatePipelineDialog } from '@/components/pipelines/CreatePipelineDialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Input } from '@/components/ui/input'
import { PipelineTableSkeleton } from '@/components/LoadingState'
import { ErrorState } from '@/components/ErrorState'
import { Search } from 'lucide-react'
import type { PipelineSummary } from '@/gen/portwhine/v1/operator_pb'

export function PipelinesPage() {
  const { data: pipelines, isLoading, isError, refetch } = usePipelines()
  const deleteMutation = useDeletePipeline()
  const [pipelineToDelete, setPipelineToDelete] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')

  const confirmDelete = () => {
    if (pipelineToDelete) {
      deleteMutation.mutate(pipelineToDelete)
      setPipelineToDelete(null)
    }
  }

  const filteredPipelines = pipelines?.filter(
    (p: PipelineSummary) =>
      p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (p.description || '').toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="flex flex-col gap-6 p-8 animate-fade-in">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Pipelines</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage your security pipeline configurations
          </p>
        </div>
        <CreatePipelineDialog />
      </div>

      <div className="relative max-w-sm">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search pipelines..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      {isLoading ? (
        <PipelineTableSkeleton />
      ) : isError ? (
        <ErrorState
          type="server"
          title="Failed to load pipelines"
          message="Unable to fetch pipeline configurations. Please try again."
          onRetry={refetch}
        />
      ) : (
        <PipelineTable
          pipelines={filteredPipelines || []}
          onDelete={(id) => setPipelineToDelete(id)}
          searchQuery={searchQuery}
        />
      )}

      <AlertDialog
        open={!!pipelineToDelete}
        onOpenChange={() => setPipelineToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Pipeline?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. This will permanently delete the
              pipeline and all its runs.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
