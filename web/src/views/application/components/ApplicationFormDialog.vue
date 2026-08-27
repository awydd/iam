<script setup lang="ts">
import { fetchApplicationInfo } from '@/api/modules/application'
import type {
  ApplicationCreateReq,
  ApplicationListItemResp,
  ApplicationUpdateInfoReq,
} from '@/api/types/application'
import { message, reportError } from '@/utils/message'
import type { FormInstance, FormRules } from 'element-plus'
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  mode: 'create' | 'edit'
  row?: ApplicationListItemResp | null
}>()

const visible = defineModel<boolean>('visible', { default: false })
const submitting = defineModel<boolean>('submitting', { default: false })

const emit = defineEmits<{
  submit: [payload: ApplicationCreateReq | ApplicationUpdateInfoReq, id?: number]
}>()

const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  client_id: '',
  redirect_uris: '',
  type: 'confidential',
  status: 'active',
  access_token_ttl: 900,
  refresh_token_ttl: 604800,
})

function resetForm() {
  form.name = ''
  form.client_id = ''
  form.redirect_uris = ''
  form.type = 'confidential'
  form.status = 'active'
  form.access_token_ttl = 900
  form.refresh_token_ttl = 604800
  formRef.value?.clearValidate()
}

watch(
  () => props.row,
  async row => {
    if (props.mode === 'edit' && row) {
      try {
        const { data } = await fetchApplicationInfo(row.id)
        if (data.status) {
          form.name = data.data.name
          form.client_id = data.data.client_id
          form.redirect_uris = Array.isArray(data.data.redirect_uris)
            ? data.data.redirect_uris.join('\n')
            : data.data.redirect_uris || ''
        } else {
          message.error(data.message)
        }
      } catch (error) {
        reportError(error)
      }
    }
  },
  { immediate: true }
)

const rules: FormRules = {
  name: [{ required: true, message: t('application.form.nameRequired'), trigger: 'blur' }],
  client_id: [{ required: true, message: t('application.form.clientIdRequired'), trigger: 'blur' }],
  redirect_uris: [
    { required: true, message: t('application.form.redirectUrisRequired'), trigger: 'blur' },
  ],
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  const uris = form.redirect_uris
    .split('\n')
    .map(s => s.trim())
    .filter(Boolean)

  if (props.mode === 'create') {
    emit('submit', {
      name: form.name,
      client_id: form.client_id,
      redirect_uris: uris,
      type: form.type,
      status: form.status,
      access_token_ttl: form.access_token_ttl,
      refresh_token_ttl: form.refresh_token_ttl,
    })
  } else {
    emit(
      'submit',
      {
        name: form.name,
        client_id: form.client_id,
        redirect_uris: uris,
      },
      props.row?.id
    )
  }
}

function handleClosed() {
  resetForm()
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="mode === 'create' ? t('application.form.createTitle') : t('application.form.editTitle')"
    width="520px"
    @closed="handleClosed"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item :label="t('application.field.name')" prop="name">
        <el-input v-model="form.name" />
      </el-form-item>
      <el-form-item :label="t('application.field.clientId')" prop="client_id">
        <el-input v-model="form.client_id" />
      </el-form-item>
      <el-form-item :label="t('application.field.redirectUris')" prop="redirect_uris">
        <el-input
          v-model="form.redirect_uris"
          type="textarea"
          :rows="3"
          :placeholder="t('application.form.redirectUrisPlaceholder')"
        />
      </el-form-item>

      <template v-if="mode === 'create'">
        <el-form-item :label="t('application.field.type')" prop="type">
          <el-select v-model="form.type" style="width: 100%">
            <el-option label="confidential" value="confidential" />
            <el-option label="public" value="public" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('application.field.status')" prop="status">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="active" value="active" />
            <el-option label="disabled" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('application.field.accessTokenTtl')" prop="access_token_ttl">
          <el-input-number
            v-model="form.access_token_ttl"
            :min="60"
            :max="900"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="t('application.field.refreshTokenTtl')" prop="refresh_token_ttl">
          <el-input-number
            v-model="form.refresh_token_ttl"
            :min="3600"
            :max="604800"
            style="width: 100%"
          />
        </el-form-item>
      </template>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>
