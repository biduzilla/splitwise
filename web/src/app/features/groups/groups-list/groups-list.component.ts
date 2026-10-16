import { DatePipe } from "@angular/common";
import { Component, DestroyRef, inject, signal } from "@angular/core";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { GroupService } from "../../../core/services/group.service";
import { Group, GroupSort } from "../../../core/models/api.types";
import { finalize } from "rxjs";
import { HttpErrorResponse } from "@angular/common/http";

@Component({
  selector: 'app-groups-list',
  standalone: true,
  imports: [RouterLink, DatePipe],
  templateUrl: './groups-list.component.html',
})
export class GroupsListComponent {
  private groupService = inject(GroupService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private destroyRef = inject(DestroyRef);

  groups = signal<Group[]>([]);
  loading = signal(true);
  error = signal<string | null>(null);
  sort = signal<GroupSort>('id');

  readonly sortOptions: { value: GroupSort; label: string }[] = [
    { value: 'id', label: 'Mais antigos' },
    { value: '-id', label: 'Mais recentes' },
    { value: 'name', label: 'Nome (A–Z)' },
    { value: '-name', label: 'Nome (Z–A)' },
  ];

  constructor() {
    this.route.queryParams
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params: any) => {
        const s = (params['sort'] as GroupSort) || 'id';
        this.sort.set(s);
        this.load(s)
      });
  }

  private load(sort: GroupSort) {
    this.loading.set(true)
    this.error.set(null)

    this.groupService
      .list({ page_size: 100, sort })
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (res) => this.groups.set(res.content),
        error: (err: HttpErrorResponse) => {
          this.groups.set([]);
          this.error.set(err.error?.message ?? 'Erro ao carregar grupos');
        },
      })
  }

  onSortChange(event: Event) {
    const value = (event.target as HTMLSelectElement).value as GroupSort;
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { sort: value },
      queryParamsHandling: 'merge',
    });
  }
}


