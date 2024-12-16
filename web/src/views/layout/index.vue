<template>
  <el-container class="layout">
    <!-- 侧边栏 -->
    <el-aside width="200px">
      <div class="logo-container">
        <Logo />
      </div>
      <el-menu
        :default-active="route.path"
        router
        class="menu"
      >
        <el-menu-item index="/">
          <el-icon><Odometer /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/nodes">
          <el-icon><Connection /></el-icon>
          <span>节点管理</span>
        </el-menu-item>
        <el-menu-item index="/rules">
          <el-icon><Filter /></el-icon>
          <span>规则管理</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主要内容 -->
    <el-container>
      <!-- 顶部导航 -->
      <el-header>
        <div class="header-left">
          <el-button
            :icon="isCollapse ? 'Expand' : 'Fold'"
            @click="isCollapse = !isCollapse"
          />
        </div>
        <div class="header-right">
          <el-switch
            v-model="isDark"
            inline-prompt
            :active-icon="Moon"
            :inactive-icon="Sunny"
            @change="toggleTheme"
          />
          <el-dropdown>
            <el-button>
              仪表盘
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>Yacd</el-dropdown-item>
                <el-dropdown-item>MetaCubeXD</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main>
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { useTheme } from '@/composables/theme'
import Logo from '@/components/Logo.vue'
import {
  Odometer,
  Connection,
  Filter,
  Setting,
  Moon,
  Sunny,
  ArrowDown
} from '@element-plus/icons-vue'

const route = useRoute()
const isCollapse = ref(false)
const { isDark, toggleTheme } = useTheme()
</script>

<style lang="scss" scoped>
.layout {
  height: 100%;

  .el-aside {
    background-color: var(--el-menu-bg-color);
    border-right: 1px solid var(--el-border-color-light);

    .logo-container {
      height: 60px;
      display: flex;
      align-items: center;
      padding: 0 20px;
      border-bottom: 1px solid var(--el-border-color-light);
    }

    .menu {
      border-right: none;
    }
  }

  .el-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 60px;
    padding: 0 20px;
    border-bottom: 1px solid var(--el-border-color-light);

    .header-right {
      display: flex;
      align-items: center;
      gap: 20px;
    }
  }

  .el-main {
    padding: 20px;
    background-color: var(--el-bg-color);
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style> 