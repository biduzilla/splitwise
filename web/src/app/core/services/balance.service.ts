import { HttpClient, HttpParams } from "@angular/common/http";
import { Injectable, inject } from "@angular/core";
import { BalanceMap, CreateSettlementInput, Paginated, Settlement, SettlementFilters, SuggestedTransfersResponse } from "../models/api.types";
import { Observable } from "rxjs";

@Injectable({ providedIn: 'root' })
export class BalanceService {
  private http = inject(HttpClient);

  getBalances(groupId: string): Observable<BalanceMap> {
    return this.http.get<BalanceMap>(`/api/balances/${groupId}/balances`);
  }

  getSuggestedTransfers(groupId: string): Observable<SuggestedTransfersResponse> {
    return this.http.get<SuggestedTransfersResponse>(
      `/api/balances/${groupId}/suggested`,
    );
  }

  createSettlement(
    input: CreateSettlementInput,
    idempotencyKey: string,
  ): Observable<Settlement> {
    return this.http.post<Settlement>("/api/settlements", input, {
      headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  listSettlements(groupId: string, filters: Omit<SettlementFilters, "group_id"> = {}): Observable<Paginated<Settlement>> {
    let p = new HttpParams().set("group_id", groupId);
    if (filters.page) p = p.set("page", filters.page);
    if (filters.page_size) p = p.set("page_size", filters.page_size);
    if (filters.sort) p = p.set("sort", filters.sort);
    return this.http.get<Paginated<Settlement>>("/api/settlements", { params: p });
  }

  deleteSettlement(id: string): Observable<void> {
    return this.http.delete<void>(`/api/settlements/${id}`);
  }
}
