<script setup lang="ts">
import type { UserCreateReq, UserListItemResp, UserUpdateReq } from '@/api/types/user'
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import UserCreateDialog from './components/UserCreateDialog.vue'
import UserUpdateDialog from './components/UserUpdateDialog.vue'
import { useUserList } from './components/useUserList.ts'
import { useUserSessions } from './components/useUserSessions.ts'
import { useUserStatusOptions } from './components/useUserStatusOptions.ts'

const { t } = useI18n()

const {
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
} = useUserList()

const { ensureLoaded } = useUserStatusOptions()
const { submitting: sessionSubmitting, forceLogout } = useUserSessions()

function toRow(row: unknown): UserListItemResp {
  return row as UserListItemResp
}

function statusTagType(status: string) {
  if (status === 'active') return 'success'
  if (status === 'pending') return 'warning'
  return 'info'
}

const createDialogVisible = ref(false)
function openCreateDialog() {
  createDialogVisible.value = true
}
async function handleCreateSubmit(payload: UserCreateReq) {
  const ok = await submitCreate(payload)
  if (ok) createDialogVisible.value = false
}

const updateDialogVisible = ref(false)
const editingRow = ref<UserListItemResp | null>(null)
function openUpdateDialog(row: UserListItemResp) {
  if (row.is_system) {
    return
  }
  editingRow.value = row
  updateDialogVisible.value = true
}
async function handleUpdateSubmit(userId: number, payload: UserUpdateReq) {
  const ok = await submitUpdate(userId, payload)
  if (ok) updateDialogVisible.value = false
}

onMounted(() => {
  ensureLoaded()
  fetchList()
})
</script>

<template>
  <div class="user-list-view">
    <el-card shadow="never">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="query.keyword"
            :placeholder="t('user.searchPlaceholder')"
            clearable
            style="width: 220px"
            @keyup.enter="searchUsers"
          />
          <el-button type="primary" @click="searchUsers">{{ t('common.search') }}</el-button>
          <el-button @click="resetSearch">{{ t('common.reset') }}</el-button>
        </div>
        <el-button type="primary" @click="openCreateDialog">
          {{ t('user.form.createTitle') }}
        </el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="userList"
        :header-cell-style="{ textAlign: 'center' }"
        :cell-style="{ textAlign: 'center' }"
        style="width: 100%"
      >
        <el-table-column
          type="index"
          :index="i => (query.page - 1) * query.per_page + i + 1"
          :label="t('common.recordNumber')"
          width="70"
        />
        <el-table-column :label="t('user.field.username')" prop="username" min-width="140" />
        <el-table-column :label="t('user.field.email')" prop="email" min-width="180" />
        <el-table-column :label="t('user.field.status')" width="120">
          <template #default="scope">
            <el-tag :type="statusTagType(toRow(scope.row).status)">
              {{ toRow(scope.row).status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="320" fixed="right" align="left">
          <template #default="scope">
            <el-tooltip
              v-if="toRow(scope.row).is_system"
              :content="t('user.systemCannotEdit')"
              placement="top"
            >
              <el-button type="primary" link disabled>{{ t('common.edit') }}</el-button>
            </el-tooltip>
            <el-button v-else type="primary" link @click="openUpdateDialog(toRow(scope.row))">
              {{ t('common.edit') }}
            </el-button>

            <el-tooltip
              v-if="toRow(scope.row).is_system"
              :content="t('user.session.systemCannotRevoke')"
              placement="top"
            >
              <el-button type="warning" link disabled>{{
                t('user.session.forceLogout')
              }}</el-button>
            </el-tooltip>
            <el-button
              v-else
              type="warning"
              link
              :loading="sessionSubmitting"
              @click="forceLogout(toRow(scope.row))"
            >
              {{ t('user.session.forceLogout') }}
            </el-button>
            <el-tooltip
              v-if="toRow(scope.row).is_system"
              :content="t('user.systemCannotDelete')"
              placement="top"
            >
              <el-button type="danger" link disabled>{{ t('common.delete') }}</el-button>
            </el-tooltip>
            <el-button v-else type="danger" link @click="removeUser(toRow(scope.row))">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        layout="total, sizes, prev, pager, next"
        :page-sizes="[10, 20, 50]"
        :current-page="query.page"
        :page-size="query.per_page"
        :total="total"
        @current-change="changePage"
        @size-change="changePageSize"
      />
    </el-card>

    <UserCreateDialog
      v-model:visible="createDialogVisible"
      v-model:submitting="submitting"
      @submit="handleCreateSubmit"
    />

    <UserUpdateDialog
      v-model:visible="updateDialogVisible"
      v-model:submitting="submitting"
      :row="editingRow"
      @submit="handleUpdateSubmit"
    />
  </div>
</template>

<style lang="scss" scoped>
.user-list-view {
  padding: 20px;

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .toolbar-left {
      display: flex;
      gap: 8px;
    }
  }

  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
