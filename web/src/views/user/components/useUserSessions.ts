import { revokeAllUserSessions } from '@/api/modules/user'
import type { UserListItemResp } from '@/api/types/user'
import { i18n } from '@/locales'
import { message, reportError } from '@/utils/message'
import { ElMessageBox } from 'element-plus'
import { ref } from 'vue'

const { t } = i18n.global

export function useUserSessions() {
  const submitting = ref(false)

  async function forceLogout(row: UserListItemResp) {
    if (row.is_system) {
      message.warning(t('user.session.systemCannotRevoke'))
      return
    }

    try {
      await ElMessageBox.confirm(
        t('user.session.forceLogoutConfirm', { name: row.username }),
        t('common.tip'),
        { type: 'warning' }
      )
    } catch {
      return
    }

    submitting.value = true
    try {
      const { data } = await revokeAllUserSessions(row.id)
      if (data.status) {
        message.success(t('user.session.forceLogoutSuccess'))
      } else {
        message.error(data.message)
      }
    } catch (error) {
      reportError(error)
    } finally {
      submitting.value = false
    }
  }

  return { submitting, forceLogout }
}
