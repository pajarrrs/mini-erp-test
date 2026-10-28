
export function validateQty(qty: string): string | null {
  const text = qty.trim()
  if (text === '') return 'Quantity is required'
  if (!/^\d+$/.test(text)) return 'Quantity must be a whole number'
  if (Number(text) <= 0) return 'Quantity must be greater than 0'
  return null
}

export interface OrderLine {
  sku: string
  qty: string
}

export function validateOrder(customer: string, lines: OrderLine[]): Record<string, string> {
  const errors: Record<string, string> = {}

  if (customer.trim() === '') errors['customer'] = 'Customer is required'
  if (lines.length === 0) errors['items'] = 'Add at least one item'

  const usedSkus: string[] = []
  lines.forEach((line, i) => {
    if (line.sku === '') {
      errors[`items[${i}].sku`] = 'Choose a product'
    } else if (usedSkus.includes(line.sku)) {
      errors[`items[${i}].sku`] = 'This product is already in the order'
    }
    usedSkus.push(line.sku)

    const qtyError = validateQty(line.qty)
    if (qtyError) errors[`items[${i}].qty`] = qtyError
  })

  return errors
}
