import { inject } from "@angular/core";
import { CanActivateFn, ActivatedRouteSnapshot, Router } from "@angular/router";
import { map, catchError, of } from "rxjs";
import { GroupService } from "../services/group.service";

export const groupMemberGuard: CanActivateFn = (
  route: ActivatedRouteSnapshot,
) => {
  const groupService = inject(GroupService);
  const router = inject(Router);

  const groupId = route.paramMap.get("groupId");
  if (!groupId) {
    router.navigate(["/groups"]);
    return false;
  }

  return groupService.getById(groupId).pipe(
    map(() => true),
    catchError((err) => {
      router.navigate(["/groups"]);
      return of(false);
    }),
  );
};
