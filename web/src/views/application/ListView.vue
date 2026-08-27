<script setup lang="ts">
import type {
  ApplicationCreateReq,
  ApplicationListItemResp,
  ApplicationUpdateInfoReq,
} from '@/api/types/application'
import { ElMessageBox } from 'element-plus'
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import ApplicationFormDialog from './components/ApplicationFormDialog.vue'
import { useApplicationList } from './components/useApplicationList.ts'

const { t } = useI18n()

const {
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
  toggleStatus,
  regenerateSecret,
  removeApplication,
} = useApplicationList()

function toRow(row: unknown): ApplicationListItemResp {
  return row as ApplicationListItemResp
}

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const activeRow = ref<ApplicationListItemResp | null>(null)

function openCreate() {
  dialogMode.value = 'create'
  activeRow.value = null
  dialogVisible.value = true
}
function openEdit(row: ApplicationListItemResp) {
  dialogMode.value = 'edit'
  activeRow.value = row
  dialogVisible.value = true
}
async function handleFormSubmit(
  payload: ApplicationCreateReq | ApplicationUpdateInfoReq,
  id?: number
) {
  const ok =
    dialogMode.value === 'create'
      ? await submitCreate(payload as ApplicationCreateReq)
      : await submitUpdate(id!, payload as ApplicationUpdateInfoReq)
  if (ok) dialogVisible.value = false
}

async function handleRegenerateSecret(row: ApplicationListItemResp) {
  const secret = await regenerateSecret(row)
  if (secret) {
    ElMessageBox.alert(secret, t('application.newSecretTitle'), {
      confirmButtonText: t('common.confirm'),
    })
  }
}

fetchList()
</script>

<template>
  <div class="application-list-view">
    <el-card shadow="never">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="query.keyword"
            :placeholder="t('application.searchPlaceholder')"
            clearable
            style="width: 220px"
            @keyup.enter="searchApplications"
          />
          <el-button type="primary" @click="searchApplications">{{ t('common.search') }}</el-button>
          <el-button @click="resetSearch">{{ t('common.reset') }}</el-button>
        </div>
        <el-button type="primary" @click="openCreate">
          {{ t('application.form.createTitle') }}
        </el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="applicationList"
        :header-cell-style="{ textAlign: 'center' }"
        :cell-style="{ textAlign: 'center' }"
        style="width: 100%"
      >
        <el-table-column
          type="index"
          :index="i => (query.page! - 1) * query.per_page! + i + 1"
          :label="t('common.recordNumber')"
          width="70"
        />
        <el-table-column :label="t('application.field.name')" prop="name" min-width="140" />
        <el-table-column
          :label="t('application.field.clientId')"
          prop="client_id"
          min-width="160"
        />
        <el-table-column :label="t('application.field.type')" width="120">
          <template #default="scope">
            {{ toRow(scope.row).type }}
          </template>
        </el-table-column>
        <el-table-column :label="t('application.field.status')" width="120">
          <template #default="scope">
            <el-switch
              :model-value="toRow(scope.row).status === 'active'"
              @change="toggleStatus(toRow(scope.row))"
            />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="280" fixed="right" align="left">
          <template #default="scope">
            <el-button type="primary" link @click="openEdit(toRow(scope.row))">
              {{ t('common.edit') }}
            </el-button>
            <el-button link @click="handleRegenerateSecret(toRow(scope.row))">
              {{ t('application.regenerateSecret') }}
            </el-button>
            <el-button type="danger" link @click="removeApplication(toRow(scope.row))">
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

    <ApplicationFormDialog
      v-model:visible="dialogVisible"
      v-model:submitting="submitting"
      :mode="dialogMode"
      :row="activeRow"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<style lang="scss" scoped>
.application-list-view {
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
