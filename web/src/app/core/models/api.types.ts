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
  errors: Record<string, string> | null;
}

/* ============ Auth ============ */
export interface OkResponse {
  ok: boolean;
}

export interface SessionResponse {
  authenticated: boolean;
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
