import { HttpClient } from "@angular/common/http";
import { Injectable, inject } from "@angular/core";
import { BalanceMap, SuggestedTransfersResponse } from "../models/api.types";

@Injectable({ providedIn: 'root' })
export class BalanceService {
  private http = inject(HttpClient);

  getBalances(groupId: string) {
    return this.http.get<BalanceMap>(`/api/settlements/groups/${groupId}/balances`);
  }

  getSuggestedTransfers(groupId: string) {
    return this.http.get<SuggestedTransfersResponse>(
      `/api/settlements/groups/${groupId}/suggested`
    );
  }
}
