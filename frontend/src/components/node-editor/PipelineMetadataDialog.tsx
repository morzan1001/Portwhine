import { useEffect, useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { PipelinePermissions } from './PipelinePermissions'

interface PipelineMetadataDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  pipelineId: string
  name: string
  description: string
  onSave: (name: string, description: string) => void
}

export function PipelineMetadataDialog({
  open,
  onOpenChange,
  pipelineId,
  name,
  description,
  onSave,
}: PipelineMetadataDialogProps) {
  const [localName, setLocalName] = useState(name)
  const [localDescription, setLocalDescription] = useState(description)

  useEffect(() => {
    if (open) {
      setLocalName(name)
      setLocalDescription(description)
    }
  }, [open, name, description])

  const handleSave = () => {
    if (!localName.trim()) return
    onSave(localName.trim(), localDescription.trim())
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Pipeline Settings</DialogTitle>
          <DialogDescription>
            Manage pipeline details and access permissions.
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="general" className="mt-2">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="general">General</TabsTrigger>
            <TabsTrigger value="permissions">Permissions</TabsTrigger>
          </TabsList>

          <TabsContent value="general" className="space-y-4 mt-4">
            <div className="space-y-2">
              <Label htmlFor="pipeline-name">Name</Label>
              <Input
                id="pipeline-name"
                value={localName}
                onChange={(e) => setLocalName(e.target.value)}
                placeholder="Pipeline name"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="pipeline-description">Description</Label>
              <Textarea
                id="pipeline-description"
                value={localDescription}
                onChange={(e) => setLocalDescription(e.target.value)}
                placeholder="What does this pipeline do?"
                className="min-h-[100px]"
              />
            </div>

            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button onClick={handleSave} disabled={!localName.trim()}>
                Save
              </Button>
            </div>
          </TabsContent>

          <TabsContent value="permissions" className="mt-4">
            <PipelinePermissions pipelineId={pipelineId} />
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}
