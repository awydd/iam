<script setup lang="ts">
import { fetchMySessions, revokeMySession } from '@/api/modules/user'
import type { UserSessionItemResp } from '@/api/types/user'
import { reportError } from '@/utils/message'
import { asRow } from '@/utils/typeHelpers'
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const loading = ref(false)
const list = ref<UserSessionItemResp[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(10)

async function fetchList() {
  loading.value = true
  try {
    const { data } = await fetchMySessions({ page: page.value, per_page: perPage.value })
    if (data.status) {
      list.value = data.data.content
      total.value = data.data.count
    } else {
      ElMessage.error(data.message)
    }
  } catch (error) {
    reportError(error)
  } finally {
    loading.value = false
  }
}

async function handleRevoke(row: UserSessionItemResp) {
  try {
    await ElMessageBox.confirm(t('user.session.revokeConfirm'), t('common.tip'), {
      type: 'warning',
    })
  } catch {
    return
  }

  try {
    const { data } = await revokeMySession(row.session_id)
    if (data.status) {
      ElMessage.success(t('common.success'))
      fetchList()
    } else {
      ElMessage.error(data.message)
    }
  } catch (error) {
    reportError(error)
  }
}

function handlePageChange(p: number) {
  page.value = p
  fetchList()
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div v-loading="loading">
    <el-table :data="list" style="width: 100%">
      <el-table-column :label="t('user.session.ip')" prop="ip" min-width="120" />
      <el-table-column
        :label="t('user.session.device')"
        prop="user_agent"
        min-width="200"
        show-overflow-tooltip
      />
      <el-table-column
        :label="t('user.session.lastActive')"
        prop="last_active_at"
        min-width="160"
      />
      <el-table-column :label="t('user.session.createdAt')" prop="created_at" min-width="160" />
      <el-table-column :label="t('common.action')" width="140" fixed="right">
        <template #default="scope">
          <el-tag
            v-if="asRow<UserSessionItemResp>(scope.row).is_current"
            type="success"
            size="small"
          >
            {{ t('user.session.current') }}
          </el-tag>
          <el-button
            v-else
            type="danger"
            link
            @click="handleRevoke(asRow<UserSessionItemResp>(scope.row))"
          >
            {{ t('user.session.revoke') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      class="pagination"
      layout="prev, pager, next"
      :current-page="page"
      :page-size="perPage"
      :total="total"
      @current-change="handlePageChange"
    />
  </div>
</template>

<style lang="scss" scoped>
.pagination {
  margin-top: 16px;
}
</style>
