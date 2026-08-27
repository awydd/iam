import { fetchTokenList, revokeToken } from '@/api/modules/token'
import type { TokenListItemResp, TokenListReq } from '@/api/types/token'
import { i18n } from '@/locales'
import { message, reportError } from '@/utils/message'
import { ElMessageBox } from 'element-plus'
import { reactive, ref } from 'vue'

const { t } = i18n.global

export function useTokenList() {
  const loading = ref(false)

  const tokenList = ref<TokenListItemResp[]>([])
  const total = ref(0)

  const query = reactive<TokenListReq>({
    page: 1,
    per_page: 10,
  })

  async function fetchList() {
    loading.value = true
    try {
      const { data } = await fetchTokenList(query)
      if (data.status) {
        tokenList.value = data.data.content
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

  async function submitRevoke(row: TokenListItemResp) {
    try {
      await ElMessageBox.confirm(t('token.revokeConfirm'), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return
    }

    try {
      const { data } = await revokeToken(row.id)
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
    tokenList,
    total,
    query,
    fetchList,
    changePage,
    changePageSize,
    submitRevoke,
  }
}
