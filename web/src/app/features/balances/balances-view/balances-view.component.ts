import { CurrencyPipe } from "@angular/common";
import { Component, OnInit, computed, inject, signal } from "@angular/core";
import { RouterLink, ActivatedRoute } from "@angular/router";
import { GroupContextService } from "../../../core/context/group-context.service";
import { AuthService } from "../../../core/services/auth.service";
import { BalanceService } from "../../../core/services/balance.service";
import { BalanceMap, SuggestedTransfer } from "../../../core/models/api.types";
import { finalize } from "rxjs";

@Component({
  selector: 'app-balances-view',
  standalone: true,
  imports: [CurrencyPipe, RouterLink],
  templateUrl: './balances-view.component.html',
})
export class BalancesViewComponent implements OnInit {
  private balanceService = inject(BalanceService);
  private route = inject(ActivatedRoute);
  private auth = inject(AuthService);
  protected ctx = inject(GroupContextService);

  balances = signal<BalanceMap>({});
  transfers = signal<SuggestedTransfer[]>([]);
  loading = signal(true);
  error = signal<string | null>(null);

  readonly groupId = computed(
    () => this.route.snapshot.paramMap.get("groupId")
  )

  readonly currentUserId = this.auth.currentUserId

  readonly myBalance = computed(() => {
    const me = this.currentUserId()
    if (!me) return 0
    return this.balances()[me] ?? 0
  })

  ngOnInit() {
    this.load();
  }

  private load() {
    const groupId = this.groupId()
    if (!groupId) return

    this.loading.set(true)
    this.error.set(null)

    this.balanceService.getBalances(groupId)
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (balances) => {
          this.balances.set(balances)
          this.loadSuggested(groupId)
        },
        error: (err) => {
          this.error.set(err.error?.message ?? 'Erro ao carregar saldos');
          this.loading.set(false);
        }
      })
  }

  private loadSuggested(groupId: string) {
    this.loading.set(true)
    this.balanceService
      .getSuggestedTransfers(groupId)
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (res) => {
          this.transfers.set(res.transfers);
        },
        error: (err) => {
          this.error.set(err.error?.message);
        }
      })
  }

  memberName(userId: string | null | undefined): string {
    if (!userId) return "Desconhecido";

    const member = this.ctx.memberMap().get(userId);
    if (member?.name) return member.name;

    return userId.slice(0, 8) + "...";
  }

  memberInitials(userId: string | null | undefined): string {
    const name = this.memberName(userId);
    const parts = name.split(" ");

    if (parts.length === 0) return "?";
    if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();

    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }

  balanceEntries(): { userId: string; amount: number }[] {
    return Object.entries(this.balances()).map(([userId, amount]) => ({
      userId,
      amount,
    }));
  }
}
