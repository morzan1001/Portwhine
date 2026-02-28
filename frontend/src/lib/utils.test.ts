import { describe, expect, it } from 'vitest'
import { cn, timestampToDate } from './utils'

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('deduplicates tailwind classes', () => {
    expect(cn('p-4', 'p-2')).toBe('p-2')
  })

  it('handles conditional classes', () => {
    expect(cn('base', false && 'hidden', 'end')).toBe('base end')
  })
})

describe('timestampToDate', () => {
  it('returns undefined for undefined input', () => {
    expect(timestampToDate(undefined)).toBeUndefined()
  })

  it('converts a timestamp with bigint seconds', () => {
    const ts = { seconds: BigInt(1700000000), nanos: 0 }
    const result = timestampToDate(ts as any)
    expect(result).toEqual(new Date(1700000000 * 1000))
  })

  it('converts a timestamp with number seconds', () => {
    const ts = { seconds: 1700000000, nanos: 500_000_000 }
    const result = timestampToDate(ts as any)
    expect(result).toEqual(new Date(1700000000 * 1000 + 500))
  })
})
