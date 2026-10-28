
export interface Product {
  id: number
  sku: string
  name: string
  price: number
  stock: number
  is_active: boolean
  created_at: string
  last_movement_at: string | null
}

export interface ProductList {
  data: Product[]
  page: number
  limit: number
  total: number
}

export type ProductFilter = 'all' | 'active' | 'inactive' | 'low_stock'

export type MovementType = 'in' | 'out'

export interface StockInput {
  type: MovementType
  qty: number
  note: string
}

export type OrderStatus = 'CREATED' | 'CANCELLED'

export interface OrderItem {
  sku: string
  name: string
  qty: number
  unit_price: number
}

export interface Order {
  id: number
  customer: string
  status: OrderStatus
  total: number
  created_at: string
  cancelled_at: string | null
  items?: OrderItem[]
}

export interface OrderInput {
  customer: string
  items: { sku: string; qty: number }[]
}

export interface Summary {
  total_products: number
  active_products: number
  total_stock_value: number
  low_stock_count: number
}

export interface ApiErrorBody {
  code: string
  message: string
  fields?: Record<string, string>
  sku?: string
  available?: number
}
