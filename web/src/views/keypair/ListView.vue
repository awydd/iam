<script setup lang="ts">
import type { KeypairListItemResp } from '@/api/types/keypair'
import { useI18n } from 'vue-i18n'
import { useKeypairList } from './components/useKeypairList'

const { t } = useI18n()

const {
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
} = useKeypairList()

function toRow(row: unknown): KeypairListItemResp {
  return row as KeypairListItemResp
}

function statusTagType(status: string) {
  if (status === 'active') return 'success'
  if (status === 'grace') return 'warning'
  return 'info' // retired
}

fetchList()
</script>

<template>
  <div class="keypair-list-view">
    <el-card shadow="never">
      <div class="toolbar">
        <div />
        <el-button type="primary" :loading="submitting" @click="submitRotate">
          {{ t('keypair.rotate') }}
        </el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="keypairList"
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
        <el-table-column
          :label="t('keypair.field.kid')"
          prop="kid"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column :label="t('keypair.field.algorithm')" prop="algorithm" width="120" />
        <el-table-column :label="t('keypair.field.status')" width="120">
          <template #default="scope">
            <el-tag :type="statusTagType(toRow(scope.row).status)">
              {{ toRow(scope.row).status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="200" fixed="right" align="left">
          <template #default="scope">
            <el-button
              v-if="toRow(scope.row).status === 'active'"
              link
              @click="submitDowngrade(toRow(scope.row))"
            >
              {{ t('keypair.downgrade') }}
            </el-button>
            <el-button
              v-if="toRow(scope.row).status !== 'retired'"
              type="danger"
              link
              @click="submitRetire(toRow(scope.row))"
            >
              {{ t('keypair.retire') }}
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
  </div>
</template>

<style lang="scss" scoped>
.keypair-list-view {
  padding: 20px;

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
