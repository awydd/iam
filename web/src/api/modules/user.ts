import request from '../index'
import type { ApiResponse, ListResp, StatusOption } from '../types/response'
import type {
  UserCreateReq,
  UserInfoResp,
  UserListItemResp,
  UserListReq,
  UserMeResp,
  UserPasswordReq,
  UserSessionItemResp,
  UserSessionListReq,
  UserUpdateReq,
} from '../types/user'

export function fetchMe() {
  return request.get<ApiResponse<UserMeResp>>('/auth/me')
}

export function userPassword(body: UserPasswordReq) {
  return request.put<ApiResponse>('/auth/me/password', body)
}

export function fetchMySessions(params: UserSessionListReq) {
  return request.get<ApiResponse<ListResp<UserSessionItemResp>>>('/auth/sessions', { params })
}

export function revokeMySession(sessionId: string) {
  return request.delete<ApiResponse>(`/auth/sessions/${sessionId}`)
}

export function fetchUserList(params: UserListReq) {
  return request.get<ApiResponse<ListResp<UserListItemResp>>>('/users', { params: params })
}

export function fetchUserStatusOptions() {
  return request.get<ApiResponse<StatusOption[]>>('/users/status-options')
}

export function fetchUserInfo(userId: number) {
  return request.get<ApiResponse<UserInfoResp>>(`/users/${userId}`)
}

export function createUser(body: UserCreateReq) {
  return request.post<ApiResponse>('/users', body)
}

export function updateUser(userId: number, body: UserUpdateReq) {
  return request.put<ApiResponse>(`/users/${userId}`, body)
}

export function deleteUser(userId: number) {
  return request.delete<ApiResponse>(`/users/${userId}`)
}

export function revokeAllUserSessions(userId: number) {
  return request.delete<ApiResponse>(`/users/${userId}/sessions`)
}
