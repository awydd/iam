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
