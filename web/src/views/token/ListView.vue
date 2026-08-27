<script setup lang="ts">
import type { TokenListItemResp } from '@/api/types/token'
import { useI18n } from 'vue-i18n'
import { useTokenList } from './components/useTokenList'

const { t } = useI18n()

const { loading, tokenList, total, query, fetchList, changePage, changePageSize, submitRevoke } =
  useTokenList()

function toRow(row: unknown): TokenListItemResp {
  return row as TokenListItemResp
}

fetchList()
</script>

<template>
  <div class="token-list-view">
    <el-card shadow="never">
      <el-table
        v-loading="loading"
        :data="tokenList"
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
        <el-table-column :label="t('token.field.expiresAt')" prop="expires_at" min-width="160" />
        <el-table-column :label="t('token.field.username')" prop="username" min-width="120" />
        <el-table-column
          :label="t('token.field.applicationName')"
          prop="application_name"
          min-width="140"
        />
        <el-table-column :label="t('token.field.type')" prop="type" width="100" />
        <el-table-column :label="t('token.field.ip')" prop="ip" min-width="120" />
        <el-table-column
          :label="t('token.field.userAgent')"
          prop="user_agent"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.action')" width="100" fixed="right">
          <template #default="scope">
            <el-button type="danger" link @click="submitRevoke(toRow(scope.row))">
              {{ t('token.revoke') }}
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
.token-list-view {
  padding: 20px;

  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
