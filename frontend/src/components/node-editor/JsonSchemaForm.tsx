import { useCallback } from 'react'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'

interface JsonSchemaProperty {
  type: string
  description?: string
  default?: any
  items?: { type: string }
}

interface JsonSchema {
  type: string
  properties?: Record<string, JsonSchemaProperty>
  required?: string[]
}

interface JsonSchemaFormProps {
  schema: string
  values: Record<string, any>
  onChange: (values: Record<string, any>) => void
}

export function JsonSchemaForm({ schema, values, onChange }: JsonSchemaFormProps) {
  const handleChange = useCallback(
    (key: string, value: any) => {
      onChange({ ...values, [key]: value })
    },
    [values, onChange]
  )

  let parsed: JsonSchema
  try {
    parsed = JSON.parse(schema)
  } catch {
    return (
      <p className="text-sm text-muted-foreground">Invalid configuration schema.</p>
    )
  }

  if (!parsed.properties) {
    return (
      <p className="text-sm text-muted-foreground">No configurable properties.</p>
    )
  }

  const properties = parsed.properties
  const required = new Set(parsed.required || [])

  return (
    <div className="space-y-4">
      {Object.entries(properties).map(([key, prop]) => (
        <div key={key} className="space-y-1.5">
          <Label className="text-sm">
            {key}
            {required.has(key) && <span className="text-red-500 ml-0.5">*</span>}
          </Label>

          {prop.type === 'boolean' ? (
            <div className="flex items-center gap-2">
              <Switch
                checked={values[key] ?? prop.default ?? false}
                onCheckedChange={(checked) => handleChange(key, checked)}
              />
              {prop.description && (
                <span className="text-xs text-muted-foreground">{prop.description}</span>
              )}
            </div>
          ) : prop.type === 'number' || prop.type === 'integer' ? (
            <Input
              type="number"
              value={values[key] ?? prop.default ?? ''}
              onChange={(e) => {
                const v = e.target.value
                handleChange(key, v === '' ? undefined : Number(v))
              }}
              placeholder={
                prop.default !== undefined ? `Default: ${prop.default}` : undefined
              }
            />
          ) : prop.type === 'array' && prop.items?.type === 'string' ? (
            <Textarea
              value={
                Array.isArray(values[key])
                  ? values[key].join('\n')
                  : values[key] ?? ''
              }
              onChange={(e) => {
                const lines = e.target.value
                  .split('\n')
                  .map((l: string) => l.trim())
                  .filter(Boolean)
                handleChange(key, lines.length > 0 ? lines : undefined)
              }}
              placeholder={prop.description || 'One value per line'}
              className="font-mono text-sm min-h-[80px]"
            />
          ) : prop.type === 'object' ? (
            <Textarea
              value={
                typeof values[key] === 'object'
                  ? JSON.stringify(values[key], null, 2)
                  : values[key] ?? ''
              }
              onChange={(e) => {
                try {
                  handleChange(key, JSON.parse(e.target.value))
                } catch {
                  // Keep raw string until valid JSON
                }
              }}
              placeholder={prop.description || '{}'}
              className="font-mono text-sm min-h-[80px]"
            />
          ) : (
            <Input
              value={values[key] ?? ''}
              onChange={(e) =>
                handleChange(key, e.target.value || undefined)
              }
              placeholder={
                prop.default !== undefined
                  ? `Default: ${prop.default}`
                  : prop.description || ''
              }
            />
          )}

          {prop.description && prop.type !== 'boolean' && (
            <p className="text-xs text-muted-foreground">{prop.description}</p>
          )}
        </div>
      ))}
    </div>
  )
}
