import { computed, inject, Injectable } from "@angular/core";
import { ActivatedRoute, NavigationEnd, Router } from "@angular/router";
import { GroupService } from "../services/group.service";
import { toSignal } from "@angular/core/rxjs-interop";
import { filter, map, of, switchMap } from "rxjs";
import { GroupDetail } from "../models/api.types";

@Injectable({ providedIn: "root" })
export class GroupContextService {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private groupService = inject(GroupService);

  readonly activeGroupId = toSignal(
    this.router.events.pipe(
      filter((e) => e instanceof NavigationEnd),
      map(() => this.extractGroupId()),
    ),
    { initialValue: this.extractGroupId() },
  );

  readonly activeGroup = toSignal(
    this.router.events.pipe(
      filter((e) => e instanceof NavigationEnd),
      map(() => this.extractGroupId()),
      switchMap((id) => (id ? this.groupService.getById(id) : of(null))),
    ),
    { initialValue: null as GroupDetail | null },
  );

  readonly memberMap = computed(() => {
    const group = this.activeGroup();
    const map = new Map<string, {
      name: string;
      email: string
    }>();
    if (!group) return map;
    for (const m of group.members) {
      if (m.user) {
        map.set(m.user_id, {
          name: m.user.name,
          email: m.user.email
        });
      }
    }
    return map;
  });

  private extractGroupId(): string | null {
    let current = this.route.firstChild
    while (current) {
      const id = current.snapshot?.paramMap.get("groupId")
      if (id) return id
      current = current.firstChild
    }
    return null
  }
}
