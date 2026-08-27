<script setup lang="ts">
import { fetchMe } from '@/api/modules/user'
import type { UserMeResp } from '@/api/types/user'
import { reportError } from '@/utils/message'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const userInfo = ref<UserMeResp | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    const { data } = await fetchMe()
    if (data.status) {
      userInfo.value = data.data
    }
  } catch (error) {
    reportError(error)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="dashboard-container" v-loading="loading">
    <h1 v-if="userInfo" class="greeting-title">
      {{ t('dashboard.greeting', { name: userInfo.username }) }}
    </h1>

    <h1 v-else-if="!loading" class="greeting-title">Dashboard</h1>
  </div>
</template>

<style lang="scss" scoped>
.dashboard-container {
  padding: 24px;

  .greeting-title {
    font-size: 28px;
    font-weight: 600;
    color: var(--text-color);
    margin-bottom: 20px;
  }
}
</style>
