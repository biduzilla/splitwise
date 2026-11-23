import { Component, DestroyRef, computed, inject, signal } from "@angular/core";
import { HttpErrorResponse } from "@angular/common/http";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import { CurrencyPipe } from "@angular/common";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { finalize } from "rxjs";
import {
  form,
  FormField,
  required,
  min,
} from "@angular/forms/signals";

import { BalanceService } from "../../../core/services/balance.service";
import { GroupContextService } from "../../../core/context/group-context.service";
import { CreateSettlementInput } from "../../../core/models/api.types";

interface SettlementFormData {
  from_user_id: string;
  to_user_id: string;
  amount: number | null;
}

@Component({
  selector: "app-settlement-form",
  standalone: true,
  imports: [FormField, RouterLink, CurrencyPipe],
  templateUrl: "./settlement-form.component.html",
})
export class SettlementFormComponent {
  private balanceService = inject(BalanceService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private destroyRef = inject(DestroyRef);
  protected ctx = inject(GroupContextService);

  readonly groupId = computed(
    () => this.route.snapshot.paramMap.get("groupId") ?? "",
  );

  readonly members = computed(() => this.ctx.activeGroup()?.members ?? []);

  model = signal<SettlementFormData>({
    from_user_id: "",
    to_user_id: "",
    amount: null,
  });

  settlementForm = form(this.model, (path) => {
    required(path.from_user_id, { message: "Quem pagou é obrigatório" });
    required(path.to_user_id, { message: "Quem recebeu é obrigatório" });
    required(path.amount, { message: "Valor é obrigatório" });
    min(path.amount, 1, { message: "Valor deve ser maior que zero" });
  });

  submitting = signal(false);
  error = signal<string | null>(null);
  fieldErrors = signal<Record<string, string>>({});

  readonly prefillAmount = computed(() => this.model().amount ?? 0);

  constructor() {
    this.applyPrefill();
    console.log("teste - m", this.members())
  }

  private applyPrefill() {
    const qp = this.route.snapshot.queryParamMap;
    const from = qp.get("from");
    const to = qp.get("to");
    const amount = qp.get("amount");

    if (from || to || amount) {
      this.model.update((m) => ({
        ...m,
        from_user_id: from ?? m.from_user_id,
        to_user_id: to ?? m.to_user_id,
        amount: amount ? Number(amount) : m.amount,
      }));
    }
  }

  // ── Helpers de template ──

  memberName(id: string): string {
    return this.ctx.memberMap().get(id)?.name ?? "—";
  }

  isFromToSame(): boolean {
    const m = this.model();
    return m.from_user_id !== "" && m.from_user_id === m.to_user_id;
  }

  // ── Submit ──

  onSubmit() {

    if (this.settlementForm().invalid()) return;
    if (this.isFromToSame()) {
      this.error.set("Pagador e recebedor não podem ser a mesma pessoa.");
      return;
    }

    const data = this.model();
    const groupId = this.groupId();

    const payload: CreateSettlementInput = {
      group_id: groupId,
      from_user_id: data.from_user_id,
      to_user_id: data.to_user_id,
      amount: data.amount!,
    };

    this.submitting.set(true);
    this.error.set(null);
    this.fieldErrors.set({});

    const idempotencyKey = crypto.randomUUID();

    this.balanceService
      .createSettlement(payload, idempotencyKey)
      .pipe(
        takeUntilDestroyed(this.destroyRef),
        finalize(() => this.submitting.set(false)),
      )
      .subscribe({
        next: () => this.router.navigate(["/groups", groupId, "balances"]),
        error: (err: HttpErrorResponse) => this.handleError(err),
      });
  }

  private handleError(err: HttpErrorResponse) {
    if (err.status === 422 && err.error?.errors) {
      this.fieldErrors.set(err.error.errors);
      this.error.set(err.error.message ?? "Erro de validação");
      return;
    }
    this.error.set(err.error?.message ?? "Falha ao registrar acerto");
  }

  formatCurrency(value: number | null): string {
    if (value === null || value === undefined) return '';
    const cents = value.toString().padStart(3, '0');
    const reais = cents.slice(0, -2);
    const centavos = cents.slice(-2);

    const reaisComMilhar = reais.replace(/\B(?=(\d{3})+(?!\d))/g, ".");

    return `${reaisComMilhar},${centavos}`;
  }

  onAmountInput(event: Event) {
    const input = event.target as HTMLInputElement;

    let rawValue = input.value.replace(/\D/g, '');

    if (rawValue === '') {
      this.model.update((m) => ({ ...m, amount: null }));
      input.value = '';
      return;
    }

    if (rawValue.length > 10) {
      rawValue = rawValue.slice(0, 10);
    }

    this.model.update((m) => ({ ...m, amount: parseInt(rawValue, 10) }));

    input.value = this.formatCurrency(parseInt(rawValue, 10));
  }
}
