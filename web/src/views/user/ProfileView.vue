<script setup lang="ts">
import { fetchMe } from '@/api/modules/user.ts'
import type { UserMeResp } from '@/api/types/user'
import { reportError } from '@/utils/message'
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import ChangePasswordDialog from './components/ChangePasswordDialog.vue'
import SessionList from './components/SessionList.vue'

const { t } = useI18n()
const loading = ref(false)
const passwordDialogVisible = ref(false)

const userInfo = ref<UserMeResp>({
  username: '',
  email: '',
  uuid: '',
  is_system: false,
})

async function fetchProfile() {
  loading.value = true
  try {
    const { data } = await fetchMe()
    if (data.status) {
      userInfo.value = data.data
    } else {
      ElMessage.error(data.message)
    }
  } catch (error) {
    reportError(error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchProfile()
})
</script>

<template>
  <div class="profile-container" v-loading="loading" v-if="userInfo">
    <el-row :gutter="20" class="profile-row">
      <el-col :span="8">
        <el-card class="user-card height-100" shadow="never">
          <div class="avatar-wrapper">
            <el-avatar :size="80">
              <el-icon :size="40"><EpUserFilled /></el-icon>
            </el-avatar>
            <h3>{{ userInfo.username || '—' }}</h3>
          </div>
        </el-card>
      </el-col>

      <el-col :span="16">
        <el-card class="height-100" shadow="never" :header="t('layout.profile')">
          <el-descriptions :column="1" border>
            <el-descriptions-item :label="t('user.field.email')">
              {{ userInfo.email || '—' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('user.field.uuid')">
              <span class="monospace">{{ userInfo.uuid }}</span>
            </el-descriptions-item>
          </el-descriptions>

          <el-tooltip
            v-if="userInfo.is_system"
            :content="t('user.systemCannotChangePassword')"
            placement="top"
          >
            <el-button class="change-password-btn" disabled>
              {{ t('user.password.title') }}
            </el-button>
          </el-tooltip>
          <el-button v-else class="change-password-btn" @click="passwordDialogVisible = true">
            {{ t('user.password.title') }}
          </el-button>
        </el-card>
      </el-col>
    </el-row>

    <el-row class="session-row">
      <el-col :span="24">
        <el-card shadow="never" :header="t('user.session.title')">
          <SessionList />
        </el-card>
      </el-col>
    </el-row>

    <ChangePasswordDialog v-model:visible="passwordDialogVisible" />
  </div>
</template>

<style lang="scss" scoped>
.profile-container {
  padding: 20px;

  .profile-row {
    display: flex;
    align-items: stretch;

    :deep(.el-col) {
      display: flex;
      flex-direction: column;
    }
  }

  .session-row {
    margin-top: 20px;
  }

  .height-100 {
    flex: 1;
    display: flex;
    flex-direction: column;

    :deep(.el-card__body) {
      flex: 1;
      display: flex;
      flex-direction: column;
      justify-content: center;
    }
  }

  .user-card {
    text-align: center;
    .avatar-wrapper {
      padding: 20px 0;
      h3 {
        margin: 16px 0 0;
        font-size: 20px;
        color: var(--el-text-color-primary);
      }
    }
  }

  .change-password-btn {
    margin-top: 16px;
    align-self: flex-start;
  }

  .monospace {
    font-family: monospace;
    font-size: 13px;
    color: var(--el-text-color-regular);
  }
}
</style>
