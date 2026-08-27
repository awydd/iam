import type { Pagination } from './request'

export interface ApplicationListReq extends Pagination {
  keyword?: string
}

export interface ApplicationCreateReq {
  name: string
  client_id: string
  redirect_uris: string[]
  type: string // confidential | public
  status: string // disabled | active
  access_token_ttl: number // 60-900
  refresh_token_ttl: number // 3600-604800
}

export interface ApplicationUpdateInfoReq {
  name: string
  client_id: string
  redirect_uris: string[]
}

export interface ApplicationUpdateTTLReq {
  access_token_ttl: number
  refresh_token_ttl: number
}

export interface ApplicationUpdateStatusReq {
  status: string
}

export interface ApplicationListItemResp {
  id: number
  name: string
  client_id: string
  redirect_uris: string[]
  type: string
  status: string
}

export interface ApplicationInfoResp {
  id: number
  name: string
  client_id: string
  redirect_uris: string[]
  type: string
  status: string
}

export interface ApplicationCreateResp {
  client_secret: string
}

export interface ApplicationUpdateSecretResp {
  client_secret: string
}
