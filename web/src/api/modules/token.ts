import request from '../index'
import type { ApiResponse, ListResp } from '../types/response'
import type { TokenInfoResp, TokenListItemResp, TokenListReq } from '../types/token'

export function fetchTokenList(params: TokenListReq) {
  return request.get<ApiResponse<ListResp<TokenListItemResp>>>('/tokens', { params })
}

export function fetchTokenInfo(id: number) {
  return request.get<ApiResponse<TokenInfoResp>>(`/tokens/${id}`)
}

export function revokeToken(id: number) {
  return request.delete<ApiResponse>(`/tokens/${id}`)
}
