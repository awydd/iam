import request from '../index'
import type { ApiResponse, ListResp } from '../types/response'
import type {
  UserMeResp,
  UserPasswordReq,
  UserSessionItemResp,
  UserSessionListReq,
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
