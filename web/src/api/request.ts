import { appConfig } from '@/configs'
import axios, { type AxiosInstance } from 'axios'

const request: AxiosInstance = axios.create({
  baseURL: appConfig.baseURL,
  timeout: 15000,
  withCredentials: true,
})

export default request
