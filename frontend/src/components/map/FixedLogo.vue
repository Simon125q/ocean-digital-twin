<template>
  <div class="fixed-logo-container">
    <img
      src="/obsea-logo.png"
      alt="OBSEA Logo"
      class="logo-image"
    />
  </div>
</template>

<script setup>
</script>

<style scoped>
.fixed-logo-container {
  position: fixed;
  top: 20px;
  right: 50px;
  z-index: 1000;
}

.logo-image {
  width: 250px;
  height: auto;
  border-radius: 8px;
  display: block;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
}
</style>

