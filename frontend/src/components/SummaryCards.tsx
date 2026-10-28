import { useEffect, useState } from 'react'
import { api, errorMessage } from '../api/client'
import type { Summary } from '../api/types'
import { formatRupiah } from '../utils/format'

interface Props {
  refreshKey: number
}

export function SummaryCards({ refreshKey }: Props) {
  const [summary, setSummary] = useState<Summary | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [retryCount, setRetryCount] = useState(0)

  useEffect(() => {
    let ignore = false
    setError(null)
    api
      .getSummary()
      .then((result) => {
        if (!ignore) setSummary(result)
      })
      .catch((err) => {
        if (!ignore) setError(errorMessage(err))
      })
    return () => {
      ignore = true
    }
  }, [refreshKey, retryCount])

  if (error) {
    return (
      <div className="alert">
        Summary could not be loaded: {error} <button onClick={() => setRetryCount(retryCount + 1)}>Retry</button>
      </div>
    )
  }

  return (
    <div className="cards">
      <Card label="Total products" value={summary ? String(summary.total_products) : '…'} />
      <Card label="Active products" value={summary ? String(summary.active_products) : '…'} />
      <Card label="Total stock value" value={summary ? formatRupiah(summary.total_stock_value) : '…'} />
      <Card label="Low stock (≤ 10)" value={summary ? String(summary.low_stock_count) : '…'} />
    </div>
  )
}

function Card({ label, value }: { label: string; value: string }) {
  return (
    <div className="card">
      <div className="card-label">{label}</div>
      <div className="card-value">{value}</div>
    </div>
  )
}
