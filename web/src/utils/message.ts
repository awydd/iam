import type { ApiResponse } from '@/api/types/response'
import { i18n } from '@/locales'
import { isAxiosError } from 'axios'
import { ElMessage, type MessageHandler, type MessageOptions } from 'element-plus'

type MessageType = 'success' | 'error' | 'warning' | 'info'

interface MessageConfig extends Partial<Omit<MessageOptions, 'message' | 'type'>> {
  _brand?: never
}

const DEFAULT_DURATION = 3000

let lastMessage = ''
let lastTime = 0

function show(type: MessageType, msg: string, options?: MessageConfig): MessageHandler | undefined {
  const now = Date.now()
  if (msg === lastMessage && now - lastTime < 1000) {
    return
  }
  lastMessage = msg
  lastTime = now

  return ElMessage({
    type,
    message: msg,
    duration: DEFAULT_DURATION,
    showClose: type === 'error',
    grouping: true,
    ...options,
  })
}

const { t } = i18n.global

export const message = {
  success(msg?: string, options?: MessageConfig) {
    return show('success', msg || t('common.operationSuccess'), options)
  },

  error(msg?: string, options?: MessageConfig) {
    return show('error', msg || t('common.operationFailed'), options)
  },

  warning(msg?: string, options?: MessageConfig) {
    return show('warning', msg || t('common.warning'), options)
  },

  info(msg?: string, options?: MessageConfig) {
    return show('info', msg || '', options)
  },

  networkError(options?: MessageConfig) {
    return show('error', t('common.networkError'), options)
  },
}

export function reportError(error: unknown, fallbackMsg?: string) {
  if (isAxiosError<ApiResponse>(error)) {
    if (error.code === 'ERR_CANCELED') {
      return
    }

    if (!error.response) {
      message.networkError()
      return
    }

    const backendMsg = error.response.data?.message
    if (backendMsg) {
      message.error(backendMsg)
      return
    }

    message.error(fallbackMsg || t('common.operationFailed'))
    return
  }

  if (error instanceof Error) {
    message.error(fallbackMsg || error.message)
    return
  }

  message.error(fallbackMsg || t('common.networkError'))
}

export default message
