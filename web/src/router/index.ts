import { createRouter, createWebHistory } from 'vue-router'
import { setupAuthGuard } from './guards/auth.guard'
import { staticRoutes } from './routes/static'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: staticRoutes,
})

setupAuthGuard(router)

export default router
