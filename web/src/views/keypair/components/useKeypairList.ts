import {
  downgradeKeypair,
  fetchKeypairList,
  retireKeypair,
  rotateKeypair,
} from '@/api/modules/keypair'
import type { KeypairListItemResp, KeypairListReq } from '@/api/types/keypair'
import { i18n } from '@/locales'
import { message, reportError } from '@/utils/message'
import { ElMessageBox } from 'element-plus'
import { reactive, ref } from 'vue'

const { t } = i18n.global

export function useKeypairList() {
  const loading = ref(false)
  const submitting = ref(false)

  const keypairList = ref<KeypairListItemResp[]>([])
  const total = ref(0)

  const query = reactive<KeypairListReq>({
    page: 1,
    per_page: 10,
  })

  async function fetchList() {
    loading.value = true
    try {
      const { data } = await fetchKeypairList(query)
      if (data.status) {
        keypairList.value = data.data.content
        total.value = data.data.count
      } else {
        message.error(data.message)
      }
    } catch (error) {
      reportError(error)
    } finally {
      loading.value = false
    }
  }

  function changePage(page: number) {
    query.page = page
    fetchList()
  }

  function changePageSize(perPage: number) {
    query.per_page = perPage
    query.page = 1
    fetchList()
  }

  async function submitRotate() {
    try {
      await ElMessageBox.confirm(t('keypair.rotateConfirm'), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return
    }

    submitting.value = true
    try {
      const { data } = await rotateKeypair()
      if (data.status) {
        message.success()
        await fetchList()
      } else {
        message.error(data.message)
      }
    } catch (error) {
      reportError(error)
    } finally {
      submitting.value = false
    }
  }

  async function submitDowngrade(row: KeypairListItemResp) {
    try {
      await ElMessageBox.confirm(t('keypair.downgradeConfirm'), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return
    }

    try {
      const { data } = await downgradeKeypair(row.kid)
      if (data.status) {
        message.success()
        await fetchList()
      } else {
        message.error(data.message)
      }
    } catch (error) {
      reportError(error)
    }
  }

  async function submitRetire(row: KeypairListItemResp) {
    try {
      await ElMessageBox.confirm(t('keypair.retireConfirm'), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return
    }

    try {
      const { data } = await retireKeypair(row.kid)
      if (data.status) {
        message.success()
        await fetchList()
      } else {
        message.error(data.message)
      }
    } catch (error) {
      reportError(error)
    }
  }

  return {
    loading,
    submitting,
    keypairList,
    total,
    query,
    fetchList,
    changePage,
    changePageSize,
    submitRotate,
    submitDowngrade,
    submitRetire,
  }
}
