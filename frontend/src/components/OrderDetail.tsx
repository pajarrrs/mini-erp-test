import { useEffect, useState } from 'react'
import { api, errorMessage } from '../api/client'
import type { Order } from '../api/types'
import { formatRupiah } from '../utils/format'
import { StatusBadge } from './OrdersPage'

interface Props {
  orderId: number
  onBack: () => void
  onCancelled: () => void
}

export function OrderDetail({ orderId, onBack, onCancelled }: Props) {
  const [order, setOrder] = useState<Order | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [cancelling, setCancelling] = useState(false)

  useEffect(() => {
    let ignore = false
    api
      .getOrder(orderId)
      .then((result) => {
        if (!ignore) setOrder(result)
      })
      .catch((err) => {
        if (!ignore) setError(errorMessage(err))
      })
    return () => {
      ignore = true
    }
  }, [orderId])

  async function handleCancel() {
    if (!window.confirm(`Cancel order #${orderId}? The stock will be put back.`)) return

    setCancelling(true)
    setError(null)
    try {
      const updated = await api.cancelOrder(orderId)
      setOrder(updated)
      onCancelled()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setCancelling(false)
    }
  }

  return (
    <section className="panel">
      <button onClick={onBack}>← Back to orders</button>

      {error && <div className="alert">{error}</div>}
      {!order && !error && <p>Loading order…</p>}

      {order && (
        <>
          <h3>
            Order #{order.id} – {order.customer} <StatusBadge status={order.status} />
          </h3>

          <table>
            <thead>
              <tr>
                <th>SKU</th>
                <th>Product</th>
                <th className="right">Qty</th>
                <th className="right">Unit price</th>
                <th className="right">Subtotal</th>
              </tr>
            </thead>
            <tbody>
              {order.items?.map((item) => (
                <tr key={item.sku}>
                  <td className="mono">{item.sku}</td>
                  <td>{item.name}</td>
                  <td className="right">{item.qty}</td>
                  <td className="right">{formatRupiah(item.unit_price)}</td>
                  <td className="right">{formatRupiah(item.unit_price * item.qty)}</td>
                </tr>
              ))}
              <tr>
                <td colSpan={4} className="right">
                  <b>Total (calculated by the server)</b>
                </td>
                <td className="right">
                  <b>{formatRupiah(order.total)}</b>
                </td>
              </tr>
            </tbody>
          </table>

          {order.status === 'CREATED' && (
            <button className="danger" onClick={handleCancel} disabled={cancelling}>
              {cancelling ? 'Cancelling…' : 'Cancel order'}
            </button>
          )}
        </>
      )}
    </section>
  )
}
