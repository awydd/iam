import request from '../index'
import type {
  ApplicationCreateReq,
  ApplicationCreateResp,
  ApplicationInfoResp,
  ApplicationListItemResp,
  ApplicationListReq,
  ApplicationUpdateInfoReq,
  ApplicationUpdateSecretResp,
  ApplicationUpdateStatusReq,
  ApplicationUpdateTTLReq,
} from '../types/application'
import type { ApiResponse, ListResp } from '../types/response'

export function fetchApplicationList(params: ApplicationListReq) {
  return request.get<ApiResponse<ListResp<ApplicationListItemResp>>>('/applications', { params })
}

export function fetchApplicationInfo(id: number) {
  return request.get<ApiResponse<ApplicationInfoResp>>(`/applications/${id}`)
}

export function createApplication(body: ApplicationCreateReq) {
  return request.post<ApiResponse<ApplicationCreateResp>>('/applications', body)
}

export function updateApplication(id: number, body: ApplicationUpdateInfoReq) {
  return request.put<ApiResponse>(`/applications/${id}`, body)
}

export function updateApplicationStatus(id: number, body: ApplicationUpdateStatusReq) {
  return request.put<ApiResponse>(`/applications/${id}/status`, body)
}

export function updateApplicationTTL(id: number, body: ApplicationUpdateTTLReq) {
  return request.put<ApiResponse>(`/applications/${id}/ttl`, body)
}

export function updateApplicationSecret(id: number) {
  return request.put<ApiResponse<ApplicationUpdateSecretResp>>(`/applications/${id}/secret`)
}

export function deleteApplication(id: number) {
  return request.delete<ApiResponse>(`/applications/${id}`)
}
