<template>
  <div class="logo">
    <div class="logo-icon">S</div>
    <span class="logo-text">SingDNS</span>
  </div>
</template>

<style lang="scss" scoped>
.logo {
  display: flex;
  align-items: center;
  gap: 10px;

  .logo-icon {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    background-color: var(--el-color-primary);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    font-weight: bold;
  }

  .logo-text {
    font-size: 20px;
    font-weight: bold;
    color: var(--el-text-color-primary);
  }
}
</style> 