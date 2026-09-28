import { HttpClient } from "@angular/common/http";
import { computed, inject, Injectable, signal } from "@angular/core";
import { Router } from "@angular/router";
import { catchError, finalize, Observable, of, shareReplay, tap } from "rxjs";
import { OkResponse, SessionResponse, SignUpInput, User } from "../models/api.types";

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);
  private router = inject(Router);

  private readonly _isAuthenticated = signal(false);
  readonly isAuthenticated = computed(() => this._isAuthenticated())

  private refresh$?: Observable<OkResponse>

  init(): Observable<SessionResponse> {
    return this.http.get<SessionResponse>('/api/auth/session').pipe(
      tap((r) => this._isAuthenticated.set(r.authenticated)),
      catchError(() => {
        this._isAuthenticated.set(false)
        return of({ authenticated: false })
      })
    )
  }

  login(email: string, password: string) {
    return this.http
      .post<OkResponse>('/api/auth', { email, password })
      .pipe(tap(() => this._isAuthenticated.set(true)))
  }

  signup(input: SignUpInput) {
    return this.http.post<User>('/api/users', input);
  }

  refreshToken(): Observable<OkResponse> {
    this.refresh$ ??= this.http.post<OkResponse>('/api/auth/refresh', {}).pipe(
      finalize(() => (this.refresh$ = undefined)),
      shareReplay({ bufferSize: 1, refCount: false }),
    );
    return this.refresh$;
  }

  logout() {
    this.http.post('/api/auth/logout', {}).subscribe({
      complete: () => this.redirectToLogin(),
      error: () => this.redirectToLogin(),
    });
  }

  private redirectToLogin() {
    this._isAuthenticated.set(false);
    this.router.navigate(['/login']);
  }
}
