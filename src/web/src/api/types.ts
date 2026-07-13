// Kiểu dữ liệu theo contracts/category-api.md (feature 001).

export type CategoryType = 'INCOME' | 'EXPENSE'

export interface User {
  id: string
  email: string
  display_name: string
}

export interface Household {
  id: string
  name: string
}

export interface Category {
  id: string
  household_id: string
  name: string
  type: CategoryType
  icon: string | null
  parent_id: string | null
  is_default: boolean
  is_hidden: boolean
  created_by: string | null
  created_at: string
  updated_at: string
}

export interface CategoryTree extends Category {
  children: Category[]
}

export interface Transaction {
  id: string
  household_id: string
  created_by: string
  amount: number
  type: CategoryType
  category_id: string
  description: string | null
  transaction_date: string
  category_name?: string
  created_by_name?: string
}

export interface Suggestion {
  category_id: string
  name: string
}
