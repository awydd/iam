import type { Router } from 'vue-router'

export function setupAuthGuard(router: Router) {
  router.beforeEach(to => {
    return true
  })
}
