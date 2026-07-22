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

export type AccountType = 'CASH' | 'BANK' | 'EWALLET' | 'CREDIT'

export interface Account {
  id: string
  name: string
  type: AccountType
  balance: number
}

export interface Transaction {
  id: string
  household_id: string
  created_by: string
  amount: number
  type: CategoryType
  category_id: string
  account_id: string
  description: string | null
  transaction_date: string
  updated_at: string
  category_name?: string
  account_name?: string
  created_by_name?: string
}

export interface Suggestion {
  category_id: string
  name: string
}

// Reports (feature 005) — contracts/report-api.md.
export type GroupUnit = 'day' | 'week' | 'month'

export interface CategoryBreakdown {
  category_id: string
  category_name: string
  category_hidden: boolean
  amount: number
  percent: number
}

export interface TrendPoint {
  bucket: string
  income: number
  expense: number
}

export interface ReportOverview {
  from: string
  to: string
  group_unit: GroupUnit
  income: number
  expense: number
  net: number
  category_breakdown: CategoryBreakdown[]
  trend: TrendPoint[]
}

export interface CategoryTrendPoint {
  bucket: string
  amount: number
}

export interface CategoryReport {
  category_id: string
  category_name: string
  from: string
  to: string
  group_unit: GroupUnit
  total: number
  transactions: Transaction[]
  trend: CategoryTrendPoint[]
}

// Budget (feature 003) — contracts/budget-api.md.
export type BudgetType = 'CATEGORY' | 'TOTAL'
export type PeriodType = 'MONTHLY' | 'WEEKLY' | 'ONE_TIME'
export type BudgetStatus = 'ACTIVE' | 'ENDED'
export type AlertLevel = 'THRESHOLD_80' | 'OVER_100'

export interface BudgetAlert {
  level: AlertLevel
  over_amount: number | null
  fired_at: string
}

// Overview / Dashboard (feature 004) — contracts/overview-api.md.
export interface MonthCashflow {
  income: number
  expense: number
  net: number
}

export interface CategorySpending {
  category_id: string
  category_name: string
  category_hidden: boolean
  amount: number
  percent: number
}

export interface OverviewSummary {
  net_worth: number
  net_worth_change_percent: number | null
  month: MonthCashflow
  category_spending: CategorySpending[]
  recent_transactions: Transaction[]
}

export interface Budget {
  id: string
  household_id: string
  type: BudgetType
  category_id: string | null
  category_name: string | null
  category_hidden: boolean | null
  limit_amount: number
  period_type: PeriodType
  start_date: string | null
  end_date: string | null
  status: BudgetStatus
  period_key: string
  spent: number
  percent: number
  alerts: BudgetAlert[]
  created_by: string
  created_by_name: string
  updated_at: string
}
