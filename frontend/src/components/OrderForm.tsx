import { useEffect, useState } from 'react'
import { api, ApiError, errorMessage } from '../api/client'
import type { Order, Product } from '../api/types'
import { type OrderLine, validateOrder } from '../utils/validation'

interface Props {
  refreshKey: number
  onCreated: (order: Order) => void
}

interface StockProblem {
  sku: string
  available: number
}

export function OrderForm({ refreshKey, onCreated }: Props) {
  const [customer, setCustomer] = useState('')
  const [lines, setLines] = useState<OrderLine[]>([{ sku: '', qty: '1' }])
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [stockProblem, setStockProblem] = useState<StockProblem | null>(null)
  const [otherError, setOtherError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  const [products, setProducts] = useState<Product[]>([])
  useEffect(() => {
    api
      .listProducts('', 'active', 1, 100)
      .then((result) => setProducts(result.data))
      .catch((err) => setOtherError('Could not load products: ' + errorMessage(err)))
  }, [refreshKey])

  function changeLine(index: number, changes: Partial<OrderLine>) {
    setLines(lines.map((line, i) => (i === index ? { ...line, ...changes } : line)))
  }

  function addLine() {
    setLines([...lines, { sku: '', qty: '1' }])
  }

  function removeLine(index: number) {
    setLines(lines.filter((_, i) => i !== index))
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    setStockProblem(null)
    setOtherError(null)

    const formErrors = validateOrder(customer, lines)
    setErrors(formErrors)
    if (Object.keys(formErrors).length > 0) return

    setSaving(true)
    try {
      const order = await api.createOrder({
        customer: customer.trim(),
        items: lines.map((line) => ({ sku: line.sku, qty: Number(line.qty) })),
      })
      setCustomer('')
      setLines([{ sku: '', qty: '1' }])
      onCreated(order)
    } catch (err) {
      if (err instanceof ApiError && err.body.code === 'INSUFFICIENT_STOCK') {
        setStockProblem({ sku: err.body.sku ?? '', available: err.body.available ?? 0 })
      } else if (err instanceof ApiError && err.body.fields) {
        setErrors(err.body.fields)
      } else {
        setOtherError(errorMessage(err))
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <form className="panel" onSubmit={handleSubmit}>
      <h3>New order</h3>

      <label>
        Customer
        <input value={customer} onChange={(e) => setCustomer(e.target.value)} placeholder="e.g. Toko Maju Jaya" />
        {errors['customer'] && <span className="field-error">{errors['customer']}</span>}
      </label>

      {lines.map((line, i) => {
        const hasStockProblem = stockProblem !== null && stockProblem.sku === line.sku
        return (
          <div key={i} className={hasStockProblem ? 'order-line failed' : 'order-line'}>
            <select value={line.sku} onChange={(e) => changeLine(i, { sku: e.target.value })}>
              <option value="">Choose a product…</option>
              {products.map((p) => (
                <option key={p.id} value={p.sku}>
                  {p.sku} – {p.name} (stock {p.stock})
                </option>
              ))}
            </select>
            <input className="qty" value={line.qty} onChange={(e) => changeLine(i, { qty: e.target.value })} inputMode="numeric" />
            <button type="button" onClick={() => removeLine(i)} disabled={lines.length === 1}>
              Remove
            </button>

            {errors[`items[${i}].sku`] && <div className="field-error">{errors[`items[${i}].sku`]}</div>}
            {errors[`items[${i}].qty`] && <div className="field-error">{errors[`items[${i}].qty`]}</div>}
            {hasStockProblem && (
              <div className="field-error">
                Not enough stock for {stockProblem.sku}: only {stockProblem.available} available
              </div>
            )}
          </div>
        )
      })}
      {errors['items'] && <div className="field-error">{errors['items']}</div>}

      <button type="button" onClick={addLine}>
        + Add item
      </button>

      {otherError && <div className="alert">{otherError}</div>}

      <div className="buttons">
        <button type="submit" className="primary" disabled={saving}>
          {saving ? 'Creating…' : 'Create order'}
        </button>
      </div>
    </form>
  )
}
