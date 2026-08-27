import type { Pagination } from './request'

export interface UserLoginReq {
  username: string
  password: string
}

export interface UserMeResp {
  username: string
  email: string
  uuid: string
  is_system: boolean
}

export interface UserPasswordReq {
  old_password: string
  new_password: string
}

export interface UserSessionListReq extends Pagination {}

export interface UserSessionItemResp {
  session_id: string
  application_id: number
  ip: string
  user_agent: string
  created_at: string
  expires_at: string
  last_active_at: string | null
  is_current: boolean
}
