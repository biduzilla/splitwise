import { Component, inject, signal } from "@angular/core";
import { form, FormField, maxLength, minLength, required } from "@angular/forms/signals";
import { Router, RouterLink } from "@angular/router";
import { GroupService } from "../../../core/services/group.service";
import { HttpErrorResponse } from "@angular/common/http";

interface GroupFormData {
  name: string;
}

@Component({
  selector: "app-group-form",
  standalone: true,
  imports: [FormField, RouterLink],
  templateUrl: "./group-form.component.html",
})
export class GroupFormComponent {
  private groupService = inject(GroupService);
  private router = inject(Router);

  model = signal<GroupFormData>({ name: "" });

  groupForm = form(this.model, (path) => {
    required(path.name, { message: "Nome é obrigatório" });
    minLength(path.name, 3, {
      message: "Nome deve ter pelo menos 3 caracteres",
    });
    maxLength(path.name, 50, {
      message: "Nome deve ter no máximo 50 caracteres",
    });
  });

  submitting = signal(false);
  error = signal<string | null>(null);
  fieldErrors = signal<Record<string, string>>({});

  onSubmit() {
    if (this.groupForm().invalid()) return

    this.submitting.set(true)
    this.error.set(null)
    this.fieldErrors.set({})

    this.groupService.create(this.model()).subscribe({
      next: (group) => this.router.navigate(['/groups', group.id]),
      error: (err: HttpErrorResponse) => {
        this.submitting.set(false);
        if (err.status === 422 && err.error?.errors) {
          this.fieldErrors.set(err.error.errors);
          this.error.set(err.error.message ?? "Erro de validação");
        } else {
          this.error.set(err.error?.message ?? "Falha ao criar grupo");
        }
      },
    })
  }
}
