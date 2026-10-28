import { useEffect, useState } from 'react'
import { api, errorMessage } from '../api/client'
import type { Order } from '../api/types'
import { formatRupiah, timeAgo } from '../utils/format'
import { OrderForm } from './OrderForm'

interface Props {
  refreshKey: number
  onOrderCreated: (order: Order) => void
  onOpenOrder: (id: number) => void
}

export function OrdersPage({ refreshKey, onOrderCreated, onOpenOrder }: Props) {
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [retryCount, setRetryCount] = useState(0)

  useEffect(() => {
    let ignore = false
    setLoading(true)
    setError(null)
    api
      .listOrders()
      .then((result) => {
        if (!ignore) setOrders(result.data)
      })
      .catch((err) => {
        if (!ignore) setError(errorMessage(err))
      })
      .finally(() => {
        if (!ignore) setLoading(false)
      })
    return () => {
      ignore = true
    }
  }, [refreshKey, retryCount])

  return (
    <section>
      <OrderForm refreshKey={refreshKey} onCreated={onOrderCreated} />

      <h3>Recent orders</h3>
      <table>
        <thead>
          <tr>
            <th>Order</th>
            <th>Customer</th>
            <th className="right">Total</th>
            <th>Status</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {loading && (
            <tr>
              <td colSpan={5} className="message">
                Loading orders…
              </td>
            </tr>
          )}
          {!loading && error && (
            <tr>
              <td colSpan={5} className="message error">
                {error} <button onClick={() => setRetryCount(retryCount + 1)}>Retry</button>
              </td>
            </tr>
          )}
          {!loading && !error && orders.length === 0 && (
            <tr>
              <td colSpan={5} className="message">
                No orders yet.
              </td>
            </tr>
          )}
          {!loading &&
            !error &&
            orders.map((order) => (
              <tr key={order.id} className="clickable" onClick={() => onOpenOrder(order.id)}>
                <td>#{order.id}</td>
                <td>{order.customer}</td>
                <td className="right">{formatRupiah(order.total)}</td>
                <td>
                  <StatusBadge status={order.status} />
                </td>
                <td>{timeAgo(order.created_at)}</td>
              </tr>
            ))}
        </tbody>
      </table>
    </section>
  )
}

export function StatusBadge({ status }: { status: Order['status'] }) {
  return <span className={status === 'CREATED' ? 'badge green' : 'badge grey'}>{status}</span>
}
