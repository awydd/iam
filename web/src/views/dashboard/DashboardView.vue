<script setup lang="ts">
import { useUserStore } from '@/stores/user'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const userStore = useUserStore()
const { userInfo, loading } = storeToRefs(userStore)

onMounted(() => {
  if (!userInfo.value) {
    userStore.getUserInfo()
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
