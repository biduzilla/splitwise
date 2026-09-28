import { Component, inject, signal } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { form, FormField, required, email } from '@angular/forms/signals';
import { AuthService } from '../../../core/services/auth.service';

interface LoginData {
  email: string;
  password: string;
}

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormField, RouterLink],
  templateUrl: './login.component.html',
})
export class LoginComponent {
  private auth = inject(AuthService);
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  justRegistered = signal(false);
  model = signal<LoginData>({ email: '', password: '' });

  loginForm = form(this.model, (path) => {
    required(path.email, { message: 'Email é obrigatório' });
    email(path.email, { message: 'Email inválido' });
    required(path.password, { message: 'Senha é obrigatória' });
  });

  isPending = signal(false);
  error = signal<string | null>(null);

  constructor() {
    this.route.queryParams.subscribe((params) => {
      this.justRegistered.set(params['registered'] === '1');
    });
  }

  onSubmit() {
    if (this.loginForm().invalid()) return;

    const { email: userEmail, password } = this.model();
    this.isPending.set(true);
    this.error.set(null);

    this.auth.login(userEmail, password).subscribe({
      next: () => this.router.navigate(['/groups']),
      error: (err: HttpErrorResponse) => {
        this.isPending.set(false);
        this.error.set(err.error?.message ?? 'Falha no login');
      },
      complete: () => this.isPending.set(false),
    });
  }
}
