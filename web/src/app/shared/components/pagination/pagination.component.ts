import { Component, EventEmitter, Input, Output } from "@angular/core";

@Component({
  selector: "app-pagination",
  standalone: true,
  template: `
    @if (lastPage > 1) {
      <div class="flex items-center justify-center gap-2 mt-4">
        <button
          type="button"
          [disabled]="currentPage === 1"
          (click)="pageChange.emit(currentPage - 1)"
          class="px-3 py-1.5 text-sm rounded-lg border border-slate-300 disabled:opacity-40 hover:bg-slate-50"
        >
          Anterior
        </button>
        <span class="text-sm text-slate-600">
          Página {{ currentPage }} de {{ lastPage }}
        </span>
        <button
          type="button"
          [disabled]="currentPage === lastPage"
          (click)="pageChange.emit(currentPage + 1)"
          class="px-3 py-1.5 text-sm rounded-lg border border-slate-300 disabled:opacity-40 hover:bg-slate-50"
        >
          Próxima
        </button>
      </div>
    }
  `,
})
export class PaginationComponent {
  @Input({ required: true }) currentPage!: number;
  @Input({ required: true }) lastPage!: number;
  @Output() pageChange = new EventEmitter<number>();
}
