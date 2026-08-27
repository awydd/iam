import type { Pagination } from './request'

export interface TokenListReq extends Pagination {
  user_id?: number
  application_id?: number
}

export interface TokenListItemResp {
  id: number
  user_id: number
  username: string
  application_id: number
  application_name: string
  session_id: string
  type: string
  ip: string
  user_agent: string
  expires_at: string
}

export type TokenInfoResp = TokenListItemResp
