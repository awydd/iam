<script setup lang="ts">
import type { UserListItemResp, UserUpdateReq } from '@/api/types/user'
import type { FormInstance, FormRules } from 'element-plus'
import { onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUserStatusOptions } from './useUserStatusOptions'

const { t } = useI18n()
const { statusOptions, ensureLoaded } = useUserStatusOptions()

onMounted(() => {
  ensureLoaded()
})

const props = defineProps<{
  row: UserListItemResp | null
}>()

const visible = defineModel<boolean>('visible', { default: false })
const submitting = defineModel<boolean>('submitting', { default: false })

const emit = defineEmits<{
  submit: [userId: number, payload: UserUpdateReq]
}>()

const formRef = ref<FormInstance>()

const form = reactive({
  username: '',
  email: '',
  password: '',
  status: '',
})

watch(
  () => props.row,
  row => {
    if (!row) return
    form.username = row.username
    form.email = row.email
    form.status = row.status
    form.password = ''
  },
  { immediate: true }
)

function resetForm() {
  formRef.value?.clearValidate()
}

const rules: FormRules = {
  username: [{ required: true, message: t('user.form.usernameRequired'), trigger: 'blur' }],
  email: [
    { required: true, message: t('user.form.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('user.form.emailInvalid'), trigger: 'blur' },
  ],
  password: [{ min: 6, max: 18, message: t('user.password.lengthRule'), trigger: 'blur' }],
  status: [{ required: true, message: t('user.form.statusRequired'), trigger: 'change' }],
}

async function handleSubmit() {
  if (!props.row) return

  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  emit('submit', props.row.id, {
    username: form.username,
    email: form.email,
    password: form.password,
    status: form.status,
  })
}

function handleClosed() {
  resetForm()
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="t('user.form.editTitle')"
    width="480px"
    @closed="handleClosed"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
      <el-form-item :label="t('user.field.username')" prop="username">
        <el-input v-model="form.username" />
      </el-form-item>
      <el-form-item :label="t('user.field.email')" prop="email">
        <el-input v-model="form.email" />
      </el-form-item>
      <el-form-item
        :label="t('user.field.password')"
        prop="password"
        :extra="t('user.form.passwordEditHint')"
      >
        <el-input v-model="form.password" type="password" show-password />
      </el-form-item>
      <el-form-item :label="t('user.field.status')" prop="status">
        <el-select v-model="form.status" style="width: 100%">
          <el-option
            v-for="item in statusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.label"
          />
        </el-select>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>
