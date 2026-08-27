<script setup lang="ts">
import { userPassword } from '@/api/modules/user'
import { reportError } from '@/utils/message'
import { type FormInstance, type FormRules } from 'element-plus'
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

const { t } = useI18n()
const router = useRouter()

const visible = defineModel<boolean>('visible', { default: false })

const formRef = ref<FormInstance>()
const submitting = ref(false)

const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

function resetForm() {
  form.old_password = ''
  form.new_password = ''
  form.confirm_password = ''
  formRef.value?.clearValidate()
}

const rules: FormRules = {
  old_password: [
    { required: true, message: t('user.password.oldRequired'), trigger: 'blur' },
    { min: 6, max: 18, message: t('user.password.lengthRule'), trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: t('user.password.newRequired'), trigger: 'blur' },
    { min: 6, max: 18, message: t('user.password.lengthRule'), trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: t('user.password.confirmRequired'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value !== form.new_password) {
          callback(new Error(t('user.password.notMatch')))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const { data } = await userPassword({
      old_password: form.old_password,
      new_password: form.new_password,
    })
    if (data.status) {
      ElMessage.success(t('user.password.success'))
      visible.value = false
      router.replace('/login')
    } else {
      ElMessage.error(data.message)
    }
  } catch (error) {
    reportError(error)
  } finally {
    submitting.value = false
  }
}

function handleClosed() {
  resetForm()
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="t('user.password.title')"
    width="420px"
    @closed="handleClosed"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item :label="t('user.password.old')" prop="old_password">
        <el-input v-model="form.old_password" type="password" show-password />
      </el-form-item>
      <el-form-item :label="t('user.password.new')" prop="new_password">
        <el-input v-model="form.new_password" type="password" show-password />
      </el-form-item>
      <el-form-item :label="t('user.password.confirm')" prop="confirm_password">
        <el-input v-model="form.confirm_password" type="password" show-password />
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
