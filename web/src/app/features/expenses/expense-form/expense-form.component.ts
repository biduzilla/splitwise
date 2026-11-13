import { Component, inject, DestroyRef, signal, computed } from "@angular/core";
import { FormField, form, required, maxLength, min } from "@angular/forms/signals";
import { RouterLink, ActivatedRoute, Router } from "@angular/router";
import { GroupContextService } from "../../../core/context/group-context.service";
import { Membership, SplitMode } from "../../../core/models/api.types";
import { AuthService } from "../../../core/services/auth.service";
import { ExpenseService } from "../../../core/services/expense.service";
import { finalize, share } from "rxjs";
import { HttpErrorResponse } from "@angular/common/http";

interface ExpenseFormData {
  amount: number | null;
  description: string;
  split_mode: SplitMode;
  paid_by: string;
  participants: string[];
  shares: Record<string, number>;
  version?: number;
}

@Component({
  selector: "app-expense-form",
  standalone: true,
  imports: [FormField, RouterLink],
  templateUrl: "./expense-form.component.html",
})
export class ExpenseFormComponent {
  private expenseService = inject(ExpenseService);
  private auth = inject(AuthService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private destroyRef = inject(DestroyRef);
  protected ctx = inject(GroupContextService);

  expenseId = signal<string | null>(null);
  isEdit = computed(() => this.expenseId() !== null);

  model = signal<ExpenseFormData>({
    amount: null,
    description: "",
    split_mode: "equal",
    paid_by: "",
    participants: [],
    shares: {},
  });

  expenseForm = form(this.model, (path) => {
    required(path.amount, { message: "Valor é obrigatório" });
    min(path.amount, 1, { message: "Valor deve ser maior que zero" });
    required(path.description, { message: "Descrição é obrigatória" });
    maxLength(path.description, 200, { message: "Máximo de 200 caracteres" });
  });

  loading = signal(false);
  submitting = signal(false);
  error = signal<string | null>(null);
  fieldErrors = signal<Record<string, string>>({});

  readonly members = computed(() => this.ctx.activeGroup()?.members ?? []);

  constructor() {
    const groupId = this.route.snapshot.paramMap.get("groupId")
    const me = this.auth.currentUserId()
    if (me) {
      this.model.update((m) => ({ ...m, paid_by: me }))
    }

    const id = this.route.snapshot.paramMap.get("id")
    this.expenseId.set(id)
    if (id) this.loadExpense(id)
  }

  onSubmit() {
    if (this.expenseForm().invalid()) return

    const data = this.model()
    const groupId = this.route.snapshot.paramMap.get("groupId")!
    const id = this.expenseId()

    this.submitting.set(true)
    this.error.set(null)
    this.fieldErrors.set({})

    const payload = this.buildPayload(groupId, data)

    const request$ = id ?
      this.expenseService.update(id, { ...payload, version: data.version! })
      : this.expenseService.create(payload, crypto.randomUUID())

    request$
      .pipe(finalize(() => this.submitting.set(false)))
      .subscribe({
        next: () => this.router.navigate(["/groups", groupId, "expenses"]),
        error: (err: HttpErrorResponse) => this.handleError(err),
      })
  }

  private buildPayload(groupId: string, data: ExpenseFormData) {
    if (data.split_mode === "equal") {
      return {
        group_id: groupId,
        paid_by: data.paid_by,
        amount: data.amount!,
        description: data.description,
        split_mode: "equal" as const,
        participants: data.participants,
      };
    }
    return {
      group_id: groupId,
      paid_by: data.paid_by,
      amount: data.amount!,
      description: data.description,
      split_mode: "exact" as const,
      shares: Object.entries(data.shares).map(([user_id, amount]) => ({
        user_id,
        amount,
      })),
    };
  }

  private loadExpense(id: string) {
    this.loading.set(true)
    this.expenseService.getById(id)
      .subscribe({
        next: (expense) => {
          const sharesMap: Record<string, number> = {}
          for (const s of expense.shares) {
            sharesMap[s.user_id] = s.amount
          }

          this.model.set({
            amount: expense.amount,
            description: expense.description,
            split_mode: expense.split_mode,
            paid_by: expense.paid_by,
            participants: expense.shares.map((s) => s.user_id),
            shares: sharesMap,
            version: expense.version,
          });
          this.loading.set(false)
        },
        error: (err) => {
          this.error.set(err.error?.message ?? "Erro ao carregar despesa");
          this.loading.set(false);
        },
      })
  }

  toggleParticipant(userId: string) {
    this.model.update((m) => {
      const has = m.participants.includes(userId)
      return {
        ...m,
        participants: has
          ? m.participants.filter((id) => id !== userId)
          : [...m.participants, userId]
      }
    })
  }

  updateShare(userId: string, cents: number) {
    this.model.update((m) => ({
      ...m,
      shares: { ...m.shares, [userId]: cents },
    }));
  }

  setSplitMode(mode: SplitMode) {
    this.model.update((m) => ({ ...m, split_mode: mode }));
  }

  readonly sharesSum = computed(() => {
    const m = this.model();
    return Object.values(m.shares).reduce((s, v) => s + (v || 0), 0);
  });

  readonly sharesMatchTotal = computed(() => {
    const m = this.model();
    if (m.split_mode !== "exact") return true;
    return this.sharesSum() === (m.amount ?? 0);
  });

  readonly canSubmit = computed(() => {
    const m = this.model();
    if (m.split_mode === "equal") return m.participants.length >= 1;
    return this.sharesMatchTotal();
  });

  private handleError(err: HttpErrorResponse) {
    this.submitting.set(false);

    if (err.status === 409) {
      this.error.set("Esta despesa foi alterada em outro lugar. Recarregue.");
      return;
    }
    if (err.status === 422 && err.error?.errors) {
      this.fieldErrors.set(err.error.errors);
      this.error.set(err.error.message ?? "Erro de validação");
      return;
    }
    this.error.set(err.error?.message ?? "Falha ao salvar despesa");
  }

  memberName(member: Membership): string {
    return member.user?.name ?? member.user_id.slice(0, 8);
  }

  changePayer(userId: string) {
    this.model.update(m => {
      const newParticipants = m.participants.includes(userId)
        ? m.participants
        : [...m.participants, userId];

      return { ...m, paid_by: userId, participants: newParticipants };
    });
  }
}
