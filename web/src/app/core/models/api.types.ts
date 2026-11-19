/* ============ Envelope ============ */
export interface Paginated<T> {
  content: T[];
  metadata: PaginationMetadata;
}

export interface PaginationMetadata {
  current_page: number;
  page_size: number;
  first_page: number;
  last_page: number;
  total_records: number;
}

/* ============ Erros ============ */
export interface ApiErrorBody {
  path: string;
  status: string;
  message: string;
  errors: Record<string, any> | null;
}

/* ============ Auth ============ */
export interface OkResponse {
  ok: boolean;
}

export interface SessionResponse {
  authenticated: boolean;
  user_id?: string;
  email?: string;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface SignUpInput {
  email: string;
  password: string;
  name: string;
}

export interface User {
  id: string;
  email: string;
  name: string;
  version: number;
}

/* ============ Group ============ */
export type GroupRole = "owner" | "member";

export type GroupSort = "id" | "-id" | "name" | "-name";

export interface Group {
  id: string;
  name: string;
  currency: string;
  owner_id: string;
  created_at: string;
  version: number;
}

export interface Membership {
  id: string;
  group_id: string;
  name: string;
  user_id: string;
  role: GroupRole;
  joined_at: string;
  user?: UserSummary;
}

export interface UserSummary {
  name: string;
  email: string;
}

export interface GroupDetail extends Group {
  members: Membership[];
}

export interface GroupFilters {
  page?: number;
  page_size?: number;
  sort?: GroupSort;
}

export interface CreateGroupInput {
  name: string;
}

export interface Invitation {
  token: string;
  url: string;
  expires_at: string;
}

/* ============ Expense ============ */
export type SplitMode = "equal" | "exact";

export interface ExpenseShare {
  user_id: string;
  amount: number;
}

export interface Expense {
  id: string;
  group_id: string;
  paid_by: string;
  amount: number;
  description: string;
  category?: string | null;
  split_mode: SplitMode;
  expense_date: string;
  version: number;
  created_at: string;
  shares: ExpenseShare[];
}

export interface CreateExpenseInput {
  group_id: string;
  paid_by: string;
  amount: number;
  description: string;
  category?: string;
  split_mode: SplitMode;
  participants?: string[];
  shares?: ExpenseShare[];
}

export interface UpdateExpenseInput extends Omit<CreateExpenseInput, 'group_id'> {
  version: number;
}

export type ExpenseSort = "id" | "-id" | "amount" | "-amount" | "expense_date" | "-expense_date";

export interface ExpenseFilters {
  page?: number;
  page_size?: number;
  sort?: ExpenseSort;
  start_date?: string;
  end_date?: string;
  user_id?: string;
}

/* ============ Balance ============ */
export interface BalanceMap {
  [userId: string]: number;
}

export interface SuggestedTransfer {
  from_user_id: string;
  to_user_id: string;
  amount: number;
}

export interface SuggestedTransfersResponse {
  transfers: SuggestedTransfer[];
}

/* ============ Settlement ============ */
// DTO retornado pelo backend em:
//   - POST /v1/settlements  (201 Created)
//   - GET  /v1/settlements  (item dentro de Paginated<Settlement>)
//
// Bate 1:1 com o SettlementDTO do Go (settlement/model.go).
export interface Settlement {
  id: string;
  group_id: string;
  from_user_id: string;
  to_user_id: string;
  amount: number;            // centavos
  settled_at: string;        // ISO 8601
  version: number;
}

// Payload de POST /v1/settlements — espelha CreateSettlementDTO do Go.
export interface CreateSettlementInput {
  group_id: string;
  from_user_id: string;
  to_user_id: string;
  amount: number;            // centavos
}

export type SettlementSort = "id" | "-id" | "settled_at" | "-settled_at";

export interface SettlementFilters {
  group_id: string;          // obrigatório no backend
  page?: number;
  page_size?: number;
  sort?: SettlementSort;
}

// Query params opcionais suportados pela rota settlements/new para
// pré-preencher o form a partir de uma sugestão clicada em
// balances-view (RN: fluxo "Registrar acerto" da sugestão).
export interface SettlementPrefill {
  from?: string;
  to?: string;
  amount?: number;
}
