import { Component, DestroyRef, inject, signal, OnInit } from "@angular/core";
import { ActivatedRoute, Router } from "@angular/router";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { GroupService } from "../../../core/services/group.service";
import { AuthService } from "../../../core/services/auth.service";

@Component({
  selector: "app-accept-invite",
  standalone: true,
  template: `
    <div class="min-h-screen flex items-center justify-center bg-slate-50 p-4">
      <div
        class="rounded-2xl border border-slate-200 bg-white p-8 shadow-xl max-w-md w-full text-center"
      >
        @if (status() === "loading") {
          <p class="text-sm text-slate-500">Aceitando convite…</p>
        } @else if (status() === "success") {
          <p class="text-sm text-emerald-600">
            Convite aceito! Redirecionando…
          </p>
        } @else if (status() === "error") {
          <p class="text-sm text-red-600">{{ errorMessage() }}</p>
          <a
            routerLink="/groups"
            class="mt-4 inline-block text-emerald-600 hover:underline"
          >
            Ir para meus grupos
          </a>
        }
      </div>
    </div>
  `,
})
export class AcceptInviteComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private groupService = inject(GroupService);
  private auth = inject(AuthService);

  status = signal<"loading" | "success" | "error">("loading");
  errorMessage = signal<string>("");

  ngOnInit() {
    if (!this.auth.isAuthenticated()) {
      const token = this.route.snapshot.queryParamMap.get("token");
      this.router.navigate(["/login"], {
        queryParams: { redirect: `/join?token=${token}` },
      });
      return;
    }

    const token = this.route.snapshot.queryParamMap.get("token");
    if (!token) {
      this.status.set("error");
      this.errorMessage.set("Token de convite ausente");
      return;
    }

    this.groupService.acceptInvitation(token).subscribe({
      next: (membership) => {
        this.status.set("success");
        setTimeout(
          () =>
            this.router.navigate(["/groups", membership.group_id, "expenses"]),
          800,
        );
      },
      error: (err) => {
        this.status.set("error");
        this.errorMessage.set(
          err.error?.message ?? "Convite inválido ou expirado",
        );
      },
    });
  }
}
