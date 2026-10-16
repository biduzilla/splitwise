import { HttpClient, HttpParams } from "@angular/common/http";
import { inject, Injectable } from "@angular/core";
import { CreateGroupInput, Group, GroupDetail, GroupFilters, Invitation, Membership, Paginated } from "../models/api.types";

@Injectable({ providedIn: 'root' })
export class GroupService {
  private http = inject(HttpClient)

  list(params: GroupFilters = {}) {
    let p = new HttpParams();
    if (params.page) p = p.set('page', params.page);
    if (params.page_size) p = p.set('page_size', params.page_size);
    if (params.sort) p = p.set('sort', params.sort);
    return this.http.get<Paginated<Group>>('/api/groups', { params: p });
  }

  getById(id: string) {
    return this.http.get<GroupDetail>(`/api/groups/${id}`);
  }

  create(input: CreateGroupInput) {
    return this.http.post<Group>('/api/groups', input);
  }

  delete(id: string) {
    return this.http.delete<void>(`/api/groups/${id}`);
  }

  createInvitation(groupId: string) {
    return this.http.post<Invitation>(`/api/groups/${groupId}/invitations`, {});
  }

  acceptInvitation(token: string) {
    return this.http.post<Membership>(`/api/invitations/${token}/accept`, {});
  }

  removeMember(groupId: string, userId: string) {
    return this.http.delete<void>(`/api/groups/${groupId}/members/${userId}`);
  }
}
