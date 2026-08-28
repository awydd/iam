<script setup lang="ts">
import Header from './components/HeaderView.vue'
import Sidebar from './components/SidebarView.vue'

const isCollapsed = ref(false)

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
}
</script>

<template>
  <div>
    <el-container>
      <!-- sidebar -->
      <Sidebar :collapsed="isCollapsed" />

      <el-container class="admin-container">
        <!-- header -->
        <Header :collapsed="isCollapsed" @toggle-collapse="toggleCollapse" />

        <el-main>
          <el-scrollbar>
            <router-view />
          </el-scrollbar>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<style lang="scss" scoped>
.admin-container {
  flex-direction: column;
  height: 100vh;
}
</style>
