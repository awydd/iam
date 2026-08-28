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

export interface UserSessionListReq extends Pagination {
  _brand?: never
}

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

export interface UserListReq extends Pagination {
  keyword?: string
}

export interface UserCreateReq {
  email: string
  username: string
  password: string
  status: string
}

export interface UserUpdateReq {
  email: string
  username: string
  password: string
  status: string
}

export interface UserListItemResp {
  id: number
  username: string
  email: string
  status: string
  is_system: boolean
}

export interface UserInfoResp {
  id: number
  username: string
  uuid: string
  email: string
}
