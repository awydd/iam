import request from '../index'
import type { KeypairListItemResp, KeypairListReq, KeypairRotateResp } from '../types/keypair'
import type { ApiResponse, ListResp } from '../types/response'

export function fetchKeypairList(params: KeypairListReq) {
  return request.get<ApiResponse<ListResp<KeypairListItemResp>>>('/keypairs', { params })
}

export function rotateKeypair() {
  return request.post<ApiResponse<KeypairRotateResp>>('/keypairs/rotate')
}

export function downgradeKeypair(kid: string) {
  return request.put<ApiResponse>(`/keypairs/${kid}/downgrade`)
}

export function retireKeypair(kid: string) {
  return request.put<ApiResponse>(`/keypairs/${kid}/retire`)
}
