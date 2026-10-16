import { Routes } from '@angular/router';
import { groupMemberGuard } from './core/guard/group-member.guard';
import { authGuard } from './core/guard/auth.guard';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'groups' },

  {
    path: 'login',
    loadComponent: () =>
      import('./features/auth/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: 'signup',
    loadComponent: () =>
      import('./features/auth/signup/signup.component').then((m) => m.SignupComponent),
  },
  {
    path: "join",
    loadComponent: () =>
      import("./features/auth/accept-invite/accept-invite.component").then(
        (m) => m.AcceptInviteComponent,
      ),
  },
  {
    path: "",
    canActivate: [authGuard],
    loadComponent: () =>
      import("./shared/layout/shell.component").then((m) => m.ShellComponent),
    children: [
      {
        path: 'groups',
        loadComponent: () => import('./features/groups/groups-list/groups-list.component').then(m => m.GroupsListComponent),
      },
      {
        path: 'groups/new',
        loadComponent: () => import('./features/groups/group-form/group-form.component').then(m => m.GroupFormComponent),
      },
      {
        path: 'groups/:groupId',
        canActivate: [groupMemberGuard],
        children: [
          { path: '', loadComponent: () => import('./features/groups/group-detail/group-detail.component').then(m => m.GroupDetailComponent) },
        ],
      },
    ]
  },
];
