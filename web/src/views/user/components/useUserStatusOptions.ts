import { fetchUserStatusOptions } from '@/api/modules/user'
import type { StatusOption } from '@/api/types/response'
import { reportError } from '@/utils/message'
import { ref } from 'vue'

const statusOptions = ref<StatusOption[]>([])
const loaded = ref(false)
const loading = ref(false)

export function useUserStatusOptions() {
  async function ensureLoaded() {
    if (loaded.value || loading.value) return
    loading.value = true
    try {
      const { data } = await fetchUserStatusOptions()
      if (data.status) {
        statusOptions.value = data.data
        loaded.value = true
      }
    } catch (error) {
      reportError(error)
    } finally {
      loading.value = false
    }
  }

  return { statusOptions, loading, ensureLoaded }
}
