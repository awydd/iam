import type { Pagination } from './request'

export interface KeypairListReq extends Pagination {
  _brand?: never
}

export interface KeypairListItemResp {
  kid: string
  algorithm: string // RS256 | ES256
  status: string // active | grace | retired
  activated_at: string
  retire_at: string | null
}

export interface KeypairRotateResp {
  kid: string
}
