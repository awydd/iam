import request from '../index'
import type { ApiResponse } from '../types/response'
import type { UserLoginReq } from '../types/user'

export function login(data: UserLoginReq) {
  return request.post<ApiResponse>('/auth/login', data)
}

export function logout() {
  return request.post<ApiResponse>('/auth/logout')
}
