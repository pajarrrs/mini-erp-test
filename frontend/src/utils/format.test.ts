import { describe, expect, it } from 'vitest'
import { formatRupiah, timeAgo } from './format'

describe('formatRupiah', () => {
  it('adds dots between thousands', () => {
    expect(formatRupiah(1234567)).toBe('Rp 1.234.567')
    expect(formatRupiah(78500)).toBe('Rp 78.500')
    expect(formatRupiah(999)).toBe('Rp 999')
    expect(formatRupiah(0)).toBe('Rp 0')
  })
})

describe('timeAgo', () => {
  const now = new Date('2026-10-28T12:00:00Z')

  it('shows a dash when there is no date', () => {
    expect(timeAgo(null, now)).toBe('-')
  })

  it('shows minutes, hours and days', () => {
    expect(timeAgo('2026-10-28T11:59:30Z', now)).toBe('just now')
    expect(timeAgo('2026-10-28T11:55:00Z', now)).toBe('5m ago')
    expect(timeAgo('2026-10-28T10:00:00Z', now)).toBe('2h ago')
    expect(timeAgo('2026-10-25T12:00:00Z', now)).toBe('3d ago')
  })
})
