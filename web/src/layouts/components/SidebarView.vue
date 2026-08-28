<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { RouteRecordRaw } from 'vue-router'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

defineProps<{ collapsed: boolean }>()

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const userStore = useUserStore()

const menuRoutes = computed<RouteRecordRaw[]>(() => {
  const layoutRoute = router.options.routes.find(r => r.path === '/')
  const children = layoutRoute?.children ?? []
  return children.filter(r => {
    if (r.meta?.public || r.meta?.hidden) {
      return false
    }

    if (!userStore.isSystem) {
      return r.path === 'dashboard' || r.name === 'dashboard'
    }

    return true
  })
})

const activeMenu = computed(() => {
  const { meta, path } = route
  if (meta?.activeMenu) {
    return meta.activeMenu as string
  }
  return path
})
</script>

<template>
  <el-aside>
    <el-menu
      :default-active="activeMenu"
      :collapse="collapsed"
      :collapse-transition="false"
      router
      class="sidebar__menu"
    >
      <div class="sidebar__logo">
        <span class="sidebar__logo-text">IAM</span>
      </div>

      <el-menu-item
        v-for="item in menuRoutes"
        :key="item.path"
        :index="item.path.startsWith('/') ? item.path : `/${item.path}`"
      >
        <el-icon>
          <Icon :icon="`ep:${item.meta?.icon ?? 'menu'}`" />
        </el-icon>
        <template #title>
          {{ item.meta?.name ? t(`layout.menus.${item.meta.name}`) : item.name || '' }}
        </template>
      </el-menu-item>
    </el-menu>
  </el-aside>
</template>

<style lang="scss" scoped>
.el-aside {
  height: 100vh;
  width: auto;
  transition: width 0.2s ease;
  border-right: 1px solid var(--border-color);
  color: var(--text-color);

  .el-menu {
    width: 230px;
    border-right: none;
    white-space: nowrap;

    &.el-menu--collapse {
      width: 60px;
    }
  }

  .sidebar__logo {
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: visible;

    .sidebar__logo-text {
      font-size: 22px;
      font-weight: bold;
      letter-spacing: 1px;
      text-align: center;
    }
  }
}
</style>
