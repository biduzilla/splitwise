import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';

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

  // // Placeholder — Fase 3 substitui por shell + grupos
  // {
  //   path: 'groups',
  //   canActivate: [authGuard],
  //   loadComponent: () =>
  //     import('./features/groups/groups-list/groups-list.component').then(
  //       (m) => m.GroupsListComponent,
  //     ),
  // },

  // { path: '**', redirectTo: 'groups' },
];
