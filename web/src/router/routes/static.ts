import type { RouteRecordRaw } from 'vue-router'

export const staticRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    component: () => import('@/layouts/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/dashboard/DashboardView.vue'),
        meta: { icon: 'monitor', name: 'dashboard' },
      },
      {
        path: 'profile',
        name: 'profile',
        component: () => import('@/views/user/ProfileView.vue'),
        meta: { icon: 'user', name: 'profile', hidden: true },
      },
      {
        path: 'keypairs',
        name: 'keypairs',
        component: () => import('@/views/keypair/ListView.vue'),
        meta: { icon: 'key', name: 'keypair' },
      },
      {
        path: 'tokens',
        name: 'tokens',
        component: () => import('@/views/token/ListView.vue'),
        meta: { icon: 'ticket', name: 'token' },
      },
      {
        path: 'applications',
        name: 'applications',
        component: () => import('@/views/application/ListView.vue'),
        meta: { icon: 'grid', name: 'application' },
      },
      {
        path: 'users',
        name: 'users',
        component: () => import('@/views/user/ListView.vue'),
        meta: { icon: 'user', name: 'user' },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { public: true },
  },
]
