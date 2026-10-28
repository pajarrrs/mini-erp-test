import type { ApiErrorBody, Order, OrderInput, Product, ProductFilter, ProductList, StockInput, Summary } from './types'

export class ApiError extends Error {
  status: number
  body: ApiErrorBody

  constructor(status: number, body: ApiErrorBody) {
    super(body.message)
    this.status = status
    this.body = body
  }
}

export function readErrorBody(status: number, json: unknown): ApiErrorBody {
  if (typeof json === 'object' && json !== null && 'error' in json) {
    return (json as { error: ApiErrorBody }).error
  }
  if (status >= 502 && status <= 504) {
    return { code: 'SERVER_UNAVAILABLE', message: 'Cannot reach the server. Is the backend running?' }
  }
  return { code: 'HTTP_' + status, message: 'Request failed with status ' + status }
}

async function request<T>(method: string, url: string, body?: unknown): Promise<T> {
  let response: Response
  try {
    response = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError(0, { code: 'NETWORK_ERROR', message: 'Cannot reach the server. Check your connection.' })
  }

  if (response.status === 204) {
    return undefined as T
  }

  const json: unknown = await response.json().catch(() => null)
  if (!response.ok) {
    throw new ApiError(response.status, readErrorBody(response.status, json))
  }
  return json as T
}

export function buildProductsUrl(search: string, filter: ProductFilter, page: number, limit: number): string {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) })
  if (search.trim() !== '') params.set('search', search.trim())
  if (filter !== 'all') params.set('status', filter)
  return '/api/products?' + params.toString()
}

export const api = {
  listProducts: (search: string, filter: ProductFilter, page: number, limit: number) =>
    request<ProductList>('GET', buildProductsUrl(search, filter, page, limit)),
  adjustStock: (productId: number, input: StockInput) =>
    request<Product>('POST', `/api/products/${productId}/stock`, input),
  getSummary: () => request<Summary>('GET', '/api/summary'),

  listOrders: () => request<{ data: Order[] }>('GET', '/api/orders'),
  getOrder: (id: number) => request<Order>('GET', `/api/orders/${id}`),
  createOrder: (input: OrderInput) => request<Order>('POST', '/api/orders', input),
  cancelOrder: (id: number) => request<Order>('POST', `/api/orders/${id}/cancel`),
}

export function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message
  return 'Something went wrong'
}
