import { useState } from 'react'
import { OrderDetail } from './components/OrderDetail'
import { OrdersPage } from './components/OrdersPage'
import { ProductsPage } from './components/ProductsPage'
import { SummaryCards } from './components/SummaryCards'

type Page = 'products' | 'orders'

const menu: { page: Page; label: string; icon: string }[] = [
  { page: 'products', label: 'Products', icon: 'M4 7l8-4 8 4-8 4-8-4zm0 5l8 4 8-4M4 17l8 4 8-4' },
  { page: 'orders', label: 'Orders', icon: 'M6 3h12v18l-3-2-3 2-3-2-3 2V3zm3 5h6m-6 4h6' },
]

export default function App() {
  const [page, setPage] = useState<Page>('products')
  const [openOrderId, setOpenOrderId] = useState<number | null>(null)

  const [refreshKey, setRefreshKey] = useState(0)
  function refresh() {
    setRefreshKey(refreshKey + 1)
  }

  function goTo(newPage: Page) {
    setPage(newPage)
    setOpenOrderId(null)
  }

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="brand">Mini ERP Test</div>
        <nav className="menu">
          {menu.map((item) => (
            <button
              key={item.page}
              className={page === item.page ? 'menu-item active' : 'menu-item'}
              onClick={() => goTo(item.page)}
            >
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d={item.icon} />
              </svg>
              {item.label}
            </button>
          ))}
        </nav>
      </aside>

      <main className="content">
        <h1>{page === 'products' ? 'Products' : 'Orders'}</h1>

        <SummaryCards refreshKey={refreshKey} />

        {page === 'products' && <ProductsPage refreshKey={refreshKey} onStockChanged={refresh} />}

        {page === 'orders' && openOrderId === null && (
          <OrdersPage
            refreshKey={refreshKey}
            onOpenOrder={setOpenOrderId}
            onOrderCreated={(order) => {
              refresh()
              setOpenOrderId(order.id)
            }}
          />
        )}

        {page === 'orders' && openOrderId !== null && (
          <OrderDetail orderId={openOrderId} onBack={() => setOpenOrderId(null)} onCancelled={refresh} />
        )}
      </main>
    </div>
  )
}
