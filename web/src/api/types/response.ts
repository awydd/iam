// fix Unexpected any. Specify a different type.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export interface ApiResponse<T = any> {
  code: number
  status: boolean
  message: string
  data: T
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export interface ListResp<T = any> {
  content: T[]
  count: number
}

export interface StatusOption {
  value: number
  label: string
}
