import { describe, expect, it } from 'vitest'
import { buildProductsUrl, readErrorBody } from './client'

describe('readErrorBody', () => {
  it('reads the backend error envelope', () => {
    const body = readErrorBody(409, { error: { code: 'INSUFFICIENT_STOCK', message: 'x', sku: 'RICE', available: 3 } })
    expect(body.code).toBe('INSUFFICIENT_STOCK')
    expect(body.sku).toBe('RICE')
    expect(body.available).toBe(3)
  })

  it('explains a 502 from the dev proxy (backend is down)', () => {
    expect(readErrorBody(502, null).code).toBe('SERVER_UNAVAILABLE')
  })
})

describe('buildProductsUrl', () => {
  it('leaves out an empty search and the "all" filter', () => {
    expect(buildProductsUrl('  ', 'all', 1, 10)).toBe('/api/products?page=1&limit=10')
  })

  it('adds search and status', () => {
    expect(buildProductsUrl('beras', 'low_stock', 2, 10)).toBe('/api/products?page=2&limit=10&search=beras&status=low_stock')
  })
})
