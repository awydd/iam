import { createUser, deleteUser, fetchUserList, updateUser } from '@/api/modules/user'
import type { UserCreateReq, UserListItemResp, UserListReq, UserUpdateReq } from '@/api/types/user'
import { i18n } from '@/locales'
import { message, reportError } from '@/utils/message'
import { reactive, ref } from 'vue'

const { t } = i18n.global

export function useUserList() {
  const loading = ref(false)
  const submitting = ref(false)

  const userList = ref<UserListItemResp[]>([])
  const total = ref(0)

  const query = reactive<UserListReq>({
    page: 1,
    per_page: 10,
    keyword: '',
  })

  async function fetchList() {
    loading.value = true
    try {
      const { data } = await fetchUserList(query)
      if (data.status) {
        userList.value = data.data.content
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

  function searchUsers() {
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

  async function submitCreate(body: UserCreateReq) {
    submitting.value = true
    try {
      const { data } = await createUser(body)
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

  async function submitUpdate(userId: number, body: UserUpdateReq) {
    submitting.value = true
    try {
      const { data } = await updateUser(userId, body)
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

  async function removeUser(row: UserListItemResp) {
    try {
      await ElMessageBox.confirm(t('user.deleteConfirm', { name: row.username }), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return
    }

    try {
      const { data } = await deleteUser(row.id)
      if (data.status) {
        message.success()
        if (userList.value.length === 1 && query.page > 1) {
          query.page -= 1
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
    userList,
    total,
    query,
    fetchList,
    searchUsers,
    resetSearch,
    changePage,
    changePageSize,
    submitCreate,
    submitUpdate,
    removeUser,
  }
}
