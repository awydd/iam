import {
  createApplication,
  deleteApplication,
  fetchApplicationList,
  updateApplication,
  updateApplicationSecret,
  updateApplicationStatus,
  updateApplicationTTL,
} from '@/api/modules/application'
import type {
  ApplicationCreateReq,
  ApplicationListItemResp,
  ApplicationListReq,
  ApplicationUpdateInfoReq,
  ApplicationUpdateTTLReq,
} from '@/api/types/application'
import { i18n } from '@/locales'
import { message, reportError } from '@/utils/message'
import { ElMessageBox } from 'element-plus'
import { reactive, ref } from 'vue'

const { t } = i18n.global

export function useApplicationList() {
  const loading = ref(false)
  const submitting = ref(false)

  const applicationList = ref<ApplicationListItemResp[]>([])
  const total = ref(0)

  const query = reactive<ApplicationListReq>({
    page: 1,
    per_page: 10,
    keyword: '',
  })

  async function fetchList() {
    loading.value = true
    try {
      const { data } = await fetchApplicationList(query)
      if (data.status) {
        applicationList.value = data.data.content
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

  function searchApplications() {
    query.page = 1
    fetchList()
  }

  function resetSearch() {
    query.page = 1
    query.keyword = ''
    fetchList()
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

  async function submitCreate(body: ApplicationCreateReq) {
    submitting.value = true
    try {
      const { data } = await createApplication(body)
      if (data.status) {
        message.success()
        await fetchList()
        return true
      }
      message.error(data.message)
      return false
    } catch (error) {
      reportError(error)
      return false
    } finally {
      submitting.value = false
    }
  }

  async function submitUpdate(id: number, body: ApplicationUpdateInfoReq) {
    submitting.value = true
    try {
      const { data } = await updateApplication(id, body)
      if (data.status) {
        message.success()
        await fetchList()
        return true
      }
      message.error(data.message)
      return false
    } catch (error) {
      reportError(error)
      return false
    } finally {
      submitting.value = false
    }
  }

  async function submitUpdateTTL(id: number, body: ApplicationUpdateTTLReq) {
    submitting.value = true
    try {
      const { data } = await updateApplicationTTL(id, body)
      if (data.status) {
        message.success()
        await fetchList()
        return true
      }
      message.error(data.message)
      return false
    } catch (error) {
      reportError(error)
      return false
    } finally {
      submitting.value = false
    }
  }

  async function toggleStatus(row: ApplicationListItemResp) {
    const nextStatus = row.status === 'active' ? 'disabled' : 'active'
    try {
      const { data } = await updateApplicationStatus(row.id, { status: nextStatus })
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

  async function regenerateSecret(row: ApplicationListItemResp): Promise<string | null> {
    try {
      await ElMessageBox.confirm(t('application.regenerateSecretConfirm'), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return null
    }

    try {
      const { data } = await updateApplicationSecret(row.id)
      if (data.status) {
        message.success()
        return data.data.client_secret
      }
      message.error(data.message)
      return null
    } catch (error) {
      reportError(error)
      return null
    }
  }

  async function removeApplication(row: ApplicationListItemResp) {
    try {
      await ElMessageBox.confirm(
        t('application.deleteConfirm', { name: row.name }),
        t('common.tip'),
        { type: 'warning' }
      )
    } catch {
      return
    }

    try {
      const { data } = await deleteApplication(row.id)
      if (data.status) {
        message.success()
        if (applicationList.value.length === 1 && query.page! > 1) {
          query.page!--
        }
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
    applicationList,
    total,
    query,
    fetchList,
    searchApplications,
    resetSearch,
    changePage,
    changePageSize,
    submitCreate,
    submitUpdate,
    submitUpdateTTL,
    toggleStatus,
    regenerateSecret,
    removeApplication,
  }
}
