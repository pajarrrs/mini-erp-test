import { describe, expect, it } from 'vitest'
import { validateOrder, validateQty } from './validation'

describe('validateQty', () => {
  it('accepts whole numbers above 0', () => {
    expect(validateQty('1')).toBeNull()
    expect(validateQty(' 25 ')).toBeNull()
  })

  it('rejects empty, zero, negative, decimal and text', () => {
    expect(validateQty('')).toBe('Quantity is required')
    expect(validateQty('0')).toBe('Quantity must be greater than 0')
    expect(validateQty('-3')).toBe('Quantity must be a whole number')
    expect(validateQty('1.5')).toBe('Quantity must be a whole number')
    expect(validateQty('abc')).toBe('Quantity must be a whole number')
  })
})

describe('validateOrder', () => {
  it('accepts a correct order', () => {
    expect(validateOrder('Toko A', [{ sku: 'RICE', qty: '2' }])).toEqual({})
  })

  it('needs a customer and at least one item', () => {
    const errors = validateOrder('  ', [])
    expect(errors['customer']).toBe('Customer is required')
    expect(errors['items']).toBe('Add at least one item')
  })

  it('marks duplicate products, missing products and bad quantities per line', () => {
    const errors = validateOrder('Toko A', [
      { sku: 'RICE', qty: '1' },
      { sku: 'RICE', qty: '0' },
      { sku: '', qty: '1' },
    ])
    expect(errors['items[1].sku']).toBe('This product is already in the order')
    expect(errors['items[1].qty']).toBe('Quantity must be greater than 0')
    expect(errors['items[2].sku']).toBe('Choose a product')
  })
})
