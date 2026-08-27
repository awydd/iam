import request from '../index'
import type { ApiResponse } from '../types/response'
import type { UserMeResp } from '../types/user'

export function fetchMe() {
  return request.get<ApiResponse<UserMeResp>>('/auth/me')
}
