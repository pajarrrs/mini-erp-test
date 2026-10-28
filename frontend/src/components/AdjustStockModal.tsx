import { useState } from 'react'
import { api, ApiError, errorMessage } from '../api/client'
import type { MovementType, Product } from '../api/types'
import { validateQty } from '../utils/validation'

interface Props {
  product: Product
  onClose: () => void
  onSaved: () => void
}

export function AdjustStockModal({ product, onClose, onSaved }: Props) {
  const [type, setType] = useState<MovementType>('in')
  const [qty, setQty] = useState('')
  const [note, setNote] = useState('')
  const [qtyError, setQtyError] = useState<string | null>(null)
  const [serverError, setServerError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()

    const error = validateQty(qty)
    setQtyError(error)
    if (error) return

    setSaving(true)
    setServerError(null)
    try {
      await api.adjustStock(product.id, { type, qty: Number(qty), note })
      onSaved()
    } catch (err) {
      if (err instanceof ApiError && err.body.code === 'INSUFFICIENT_STOCK') {
        setServerError(`Not enough stock. Only ${err.body.available} available.`)
      } else {
        setServerError(errorMessage(err))
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="modal-background">
      <form className="modal" onSubmit={handleSubmit}>
        <h3>Adjust stock</h3>
        <p className="muted">
          {product.sku} – {product.name} (current stock: <b>{product.stock}</b>)
        </p>

        <label>
          Direction
          <select value={type} onChange={(e) => setType(e.target.value as MovementType)}>
            <option value="in">In (add stock)</option>
            <option value="out">Out (remove stock)</option>
          </select>
        </label>

        <label>
          Quantity
          <input value={qty} onChange={(e) => setQty(e.target.value)} inputMode="numeric" autoFocus />
          {qtyError && <span className="field-error">{qtyError}</span>}
        </label>

        <label>
          Note
          <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="Optional" />
        </label>

        {serverError && <div className="alert">{serverError}</div>}

        <div className="buttons">
          <button type="button" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" className="primary" disabled={saving}>
            {saving ? 'Saving…' : 'Save'}
          </button>
        </div>
      </form>
    </div>
  )
}
