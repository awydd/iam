// fix Unexpected any. Specify a different type.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export interface ApiResponse<T = any> {
  code: number
  status: boolean
  message: string
  data: T
}
