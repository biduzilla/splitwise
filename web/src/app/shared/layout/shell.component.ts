import { Component, inject } from "@angular/core";
import { RouterLink, RouterLinkActive, RouterOutlet } from "@angular/router";
import { AuthService } from "../../core/services/auth.service";
import { GroupContextService } from "../../core/context/group-context.service";

@Component({
  selector: "app-shell",
  standalone: true,
  imports: [RouterLink, RouterLinkActive, RouterOutlet],
  template: `
    <div class="min-h-screen bg-slate-50">
      <nav class="border-b border-slate-200 bg-white">
        <div class="mx-auto flex max-w-6xl items-center gap-6 px-6 py-3">
          @if (ctx.activeGroup(); as group) {
            <a
              [routerLink]="['/groups', group.id, 'expenses']"
              routerLinkActive="text-emerald-600"
              class="text-sm font-medium text-slate-600 hover:text-slate-900"
            >
              Despesas
            </a>
            <a
              [routerLink]="['/groups', group.id, 'balances']"
              routerLinkActive="text-emerald-600"
              class="text-sm font-medium text-slate-600 hover:text-slate-900"
            >
              Saldos
            </a>
            <a
              [routerLink]="['/groups', group.id]"
              [routerLinkActiveOptions]="{ exact: true }"
              routerLinkActive="text-emerald-600"
              class="text-sm font-medium text-slate-600 hover:text-slate-900"
            >
              Membros
            </a>
            <span class="ml-2 text-sm font-semibold text-slate-900">{{
              group.name
            }}</span>
          }

          <a
            routerLink="/groups"
            routerLinkActive="text-emerald-600"
            class="text-sm font-medium text-slate-600 hover:text-slate-900"
          >
            Grupos
          </a>

          <button
            type="button"
            (click)="logout()"
            class="ml-auto text-sm font-medium text-slate-500 hover:text-slate-900"
          >
            Sair
          </button>
        </div>
      </nav>

      <main>
        <router-outlet />
      </main>
    </div>
  `,
})
export class ShellComponent {
  private auth = inject(AuthService);
  protected ctx = inject(GroupContextService);

  logout() {
    this.auth.logout();
  }
}
