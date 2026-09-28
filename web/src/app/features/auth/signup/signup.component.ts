import { Component, inject, signal } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { Router, RouterLink } from '@angular/router';
import {
  form,
  FormField,
  required,
  email,
  minLength,
  maxLength,
  pattern,
  validate,
} from '@angular/forms/signals';
import { AuthService } from '../../../core/services/auth.service';

interface SignupData {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
}

@Component({
  selector: 'app-signup',
  standalone: true,
  imports: [FormField, RouterLink],
  templateUrl: './signup.component.html',
})
export class SignupComponent {
  private auth = inject(AuthService);
  private router = inject(Router);

  model = signal<SignupData>({
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
  });

  signupForm = form(this.model, (path) => {
    required(path.name, { message: 'Nome é obrigatório' });
    minLength(path.name, 3, { message: 'Nome deve ter pelo menos 3 caracteres' });
    maxLength(path.name, 100, { message: 'Nome deve ter no máximo 100 caracteres' });

    required(path.email, { message: 'Email é obrigatório' });
    email(path.email, { message: 'Email inválido' });

    required(path.password, { message: 'Senha é obrigatória' });
    minLength(path.password, 8, { message: 'Senha deve ter pelo menos 8 caracteres' });
    pattern(path.password, /^(?=.*[A-Za-z])(?=.*\d).+$/, {
      message: 'Senha deve conter letras e números',
    });

    required(path.confirmPassword, { message: 'Confirmação é obrigatória' });
    validate(path.confirmPassword, ({ value, valueOf }) => {
      if (value() !== valueOf(path.password)) {
        return { kind: 'custom', message: 'Senhas não conferem' };
      }
      return null;
    });
  });

  isPending = signal(false);
  error = signal<string | null>(null);
  fieldErrors = signal<Record<string, string>>({});

  onSubmit() {
    if (this.signupForm().invalid()) return;

    this.isPending.set(true);
    this.error.set(null);
    this.fieldErrors.set({});

    const { name, email, password } = this.model();
    this.auth.signup({ name, email, password }).subscribe({
      next: () => this.router.navigate(['/login'], { queryParams: { registered: '1' } }),
      error: (err: HttpErrorResponse) => {
        this.isPending.set(false);
        if (err.status === 422 && err.error?.errors) {
          this.fieldErrors.set(err.error.errors);
          this.error.set(err.error.message ?? 'Erro de validação');
        } else {
          this.error.set(err.error?.message ?? 'Falha ao criar conta');
        }
      },
      complete: () => this.isPending.set(false),
    });
  }

  backendError(field: keyof SignupData): string | null {
    return this.fieldErrors()[field] ?? null;
  }
}
