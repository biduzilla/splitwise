import { CurrencyPipe, DatePipe } from "@angular/common";
import { Component, computed, DestroyRef, inject, signal } from "@angular/core";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import { PaginationComponent } from "../../../shared/components/pagination/pagination.component";
import { ExpenseService } from "../../../core/services/expense.service";
import { GroupContextService } from "../../../core/context/group-context.service";
import { Expense, PaginationMetadata, ExpenseFilters } from "../../../core/models/api.types";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { finalize } from "rxjs";
import { HttpErrorResponse } from "@angular/common/http";
import { AuthService } from "../../../core/services/auth.service";

@Component({
  selector: "app-expenses-list",
  standalone: true,
  imports: [
    RouterLink,
    CurrencyPipe,
    DatePipe,
    PaginationComponent,
  ],
  templateUrl: "./expenses-list.component.html",
})
export class ExpensesListComponent {
  private expenseService = inject(ExpenseService)
  private auth = inject(AuthService)
  private route = inject(ActivatedRoute)
  private router = inject(Router)
  private destroyRef = inject(DestroyRef)
  protected ctx = inject(GroupContextService)

  expenses = signal<Expense[]>([]);
  metadata = signal<Partial<PaginationMetadata>>({});
  loading = signal(true);
  error = signal<string | null>(null);
  filters = signal<ExpenseFilters>({});

  total = computed(() => this.metadata().total_records ?? 0);
  isEmpty = computed(() => !this.loading() && this.expenses().length === 0);

  readonly currentUserId = this.auth.currentUserId

  constructor() {
    this.route.queryParams
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params) => {
        const groupId = this.route.snapshot.paramMap.get("groupId")!;
        const parsed: ExpenseFilters = {
          page: Number(params["page"]) || 1,
          page_size: Number(params["page_size"]) || 20,
          sort: params["sort"] || "-expense_date",
          start_date: params["start_date"] || undefined,
          end_date: params["end_date"] || undefined,
          user_id: params["user_id"] || undefined,
        };
        this.filters.set(parsed)
        this.load(groupId, parsed)
      })
  }

  private load(groupId: string, filters: ExpenseFilters) {
    this.loading.set(true)
    this.error.set(null)

    this.expenseService
      .list(groupId, filters)
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (res) => {
          this.expenses.set(res.content)
          this.metadata.set(res.metadata)
        },
        error: (err: HttpErrorResponse) => {
          this.expenses.set([]);
          this.metadata.set({});
          this.error.set(err.error?.message ?? "Erro ao carregar despesas");
        }
      })
  }

  delete(e: Expense) {
    if (!confirm(`Tem certeza que deseja apagar a despesa "${e.description}"?`)) {
      return;
    }

    this.error.set(null);

    this.expenseService.delete(e.id).subscribe({
      next: () => {
        this.expenses.update((list) => list.filter((expense) => expense.id !== e.id));
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(err.error?.message ?? "Erro ao apagar despesa");
      }
    });
  }

  memberName(id: string): string {
    return this.ctx.memberMap().get(id)?.name ?? "-"
  }

  totalCents(): number {
    return this.expenses().reduce((sum, e) => sum + e.amount, 0)
  }

  onFilterChange(key: string, value: string) {
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { [key]: value || null, page: 1 },
      queryParamsHandling: "merge"
    })
  }

  onPageChange(page: number) {
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { page },
      queryParamsHandling: "merge",
    });
  }

  isOwner(e: Expense): boolean {
    const groupOwnerId = this.ctx.activeGroup()?.owner_id;
    const me = this.currentUserId();

    return groupOwnerId === me;
  }
}
