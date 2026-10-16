import { DatePipe } from "@angular/common";
import { Component, inject, signal } from "@angular/core";
import { RouterLink, ActivatedRoute, Router } from "@angular/router";
import { GroupDetail, Invitation } from "../../../core/models/api.types";
import { AuthService } from "../../../core/services/auth.service";
import { GroupService } from "../../../core/services/group.service";
import { HttpErrorResponse } from "@angular/common/http";

@Component({
  selector: "app-group-detail",
  standalone: true,
  imports: [RouterLink, DatePipe],
  templateUrl: "./group-detail.component.html",
})
export class GroupDetailComponent {
  private groupService = inject(GroupService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private auth = inject(AuthService);

  group = signal<GroupDetail | null>(null);
  invitation = signal<Invitation | null>(null);
  loading = signal(true);
  error = signal<string | null>(null);
  copied = signal(false);

  readonly currentUserId = this.auth.currentUserId

  constructor() {
    const groupId = this.route.snapshot.paramMap.get("groupId")!;
    this.loadGroup(groupId);
  }

  private loadGroup(id: string) {
    this.loading.set(false)
    this.groupService.getById(id).subscribe({
      next: (res) => {
        this.group.set(res);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(err.error?.message ?? "Erro ao carregar grupo");
        this.loading.set(false);
      },
    })
  }

  generateInvite() {
    const group = this.group();
    if (!group) return;

    this.groupService.createInvitation(group.id).subscribe({
      next: (inv) => this.invitation.set(inv),
      error: (err) =>
        this.error.set(err.error?.message ?? "Erro ao gerar convite"),
    });
  }

  copyInviteUrl() {
    const url = this.invitation()?.url

    if (url) {
      navigator.clipboard.writeText(url).then(() => {
        this.copied.set(true);
        setTimeout(() => this.copied.set(false), 2000);
      });
    }
  }

  removeMember(userId: string) {
    const group = this.group();
    if (!group) return;

    this.groupService.removeMember(group.id, userId).subscribe({
      next: () => this.loadGroup(group.id),
      error: (err) =>
        this.error.set(err.error?.message ?? "Erro ao remover membro"),
    });
  }

  isOwner(): boolean {
    const group = this.group();
    return !!group && group.owner_id === this.currentUserId();
  }
}
