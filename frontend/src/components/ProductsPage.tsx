import { useEffect, useState } from 'react'
import { api, errorMessage } from '../api/client'
import type { Product, ProductFilter } from '../api/types'
import { useDebounce } from '../hooks/useDebounce'
import { formatRupiah, timeAgo } from '../utils/format'
import { AdjustStockModal } from './AdjustStockModal'

const PAGE_SIZE = 10

interface Props {
  refreshKey: number
  onStockChanged: () => void
}

export function ProductsPage({ refreshKey, onStockChanged }: Props) {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<ProductFilter>('all')
  const [page, setPage] = useState(1)
  const debouncedSearch = useDebounce(search, 300)

  const [products, setProducts] = useState<Product[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [retryCount, setRetryCount] = useState(0)

  const [adjusting, setAdjusting] = useState<Product | null>(null)

  useEffect(() => {
    let ignore = false
    setLoading(true)
    setError(null)

    api
      .listProducts(debouncedSearch, filter, page, PAGE_SIZE)
      .then((result) => {
        if (ignore) return
        setProducts(result.data)
        setTotal(result.total)
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
  }, [debouncedSearch, filter, page, refreshKey, retryCount])

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <section>
      <div className="toolbar">
        <input
          placeholder="Search by SKU or name…"
          value={search}
          onChange={(e) => {
            setSearch(e.target.value)
            setPage(1)
          }}
        />
        <select
          value={filter}
          onChange={(e) => {
            setFilter(e.target.value as ProductFilter)
            setPage(1)
          }}
        >
          <option value="all">All</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
          <option value="low_stock">Low stock (≤ 10)</option>
        </select>
      </div>

      <table>
        <thead>
          <tr>
            <th>SKU</th>
            <th>Name</th>
            <th className="right">Price</th>
            <th className="right">Stock</th>
            <th>Status</th>
            <th>Last movement</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {loading && (
            <tr>
              <td colSpan={7} className="message">
                Loading products…
              </td>
            </tr>
          )}

          {!loading && error && (
            <tr>
              <td colSpan={7} className="message error">
                {error} <button onClick={() => setRetryCount(retryCount + 1)}>Retry</button>
              </td>
            </tr>
          )}

          {!loading && !error && products.length === 0 && (
            <tr>
              <td colSpan={7} className="message">
                No products found.
              </td>
            </tr>
          )}

          {!loading &&
            !error &&
            products.map((product) => (
              <tr key={product.id}>
                <td className="mono">{product.sku}</td>
                <td>{product.name}</td>
                <td className="right">{formatRupiah(product.price)}</td>
                <td className="right">{product.stock}</td>
                <td>
                  {product.is_active ? <span className="badge green">Active</span> : <span className="badge grey">Inactive</span>}
                  {product.stock <= 10 && <span className="badge orange">Low</span>}
                </td>
                <td>{timeAgo(product.last_movement_at)}</td>
                <td>
                  <button onClick={() => setAdjusting(product)}>Adjust stock</button>
                </td>
              </tr>
            ))}
        </tbody>
      </table>

      <div className="pager">
        <button disabled={page <= 1} onClick={() => setPage(page - 1)}>
          Previous
        </button>
        <span>
          Page {page} of {totalPages} ({total} products)
        </span>
        <button disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
          Next
        </button>
      </div>

      {adjusting && (
        <AdjustStockModal
          product={adjusting}
          onClose={() => setAdjusting(null)}
          onSaved={() => {
            setAdjusting(null)
            onStockChanged()
          }}
        />
      )}
    </section>
  )
}
