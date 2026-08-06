<template>
  <div class="data-toolbar">
    <div class="toolbar-filters">
      <slot name="filters" />
    </div>
    <div class="toolbar-actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<style scoped>
.data-toolbar {
  background: var(--color-bg-2);
  border-radius: 4px;
  padding: 14px 16px;
  margin-bottom: 14px;
  box-shadow: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 10px;
}

.toolbar-filters {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  flex: 1;
  min-width: 0;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

@media (max-width: 767px) {
  .data-toolbar {
    flex-direction: column;
    align-items: stretch;
  }
  .toolbar-actions {
    justify-content: flex-end;
  }
}
</style>
