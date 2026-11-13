import { HttpClient, HttpParams } from "@angular/common/http";
import { Injectable, inject } from "@angular/core";
import { CreateExpenseInput, Expense, ExpenseFilters, Paginated, UpdateExpenseInput } from "../models/api.types";

@Injectable({ providedIn: "root" })
export class ExpenseService {
  private http = inject(HttpClient);

  list(groupId: string, filters: ExpenseFilters = {}) {
    let p = new HttpParams().set("group_id", groupId);
    if (filters.page) p = p.set("page", filters.page);
    if (filters.page_size) p = p.set("page_size", filters.page_size);
    if (filters.sort) p = p.set("sort", filters.sort);
    if (filters.start_date) p = p.set("start_date", filters.start_date);
    if (filters.end_date) p = p.set("end_date", filters.end_date);
    if (filters.user_id) p = p.set("user_id", filters.user_id);

    return this.http.get<Paginated<Expense>>("/api/expenses", { params: p });
  }

  getById(id: string) {
    return this.http.get<Expense>(`/api/expenses/${id}`);
  }

  create(input: CreateExpenseInput, idempotencyKey: string) {
    return this.http.post<Expense>("/api/expenses", input, {
      headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  update(id: string, input: UpdateExpenseInput) {
    return this.http.put<Expense>(`/api/expenses/${id}`, input);
  }

  delete(id: string) {
    return this.http.delete<void>(`/api/expenses/${id}`);
  }
}
