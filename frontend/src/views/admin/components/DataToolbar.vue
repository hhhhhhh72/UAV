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
  background: var(--bg-primary, #fff);
  border-radius: var(--card-radius, 12px);
  padding: 14px 16px;
  margin-bottom: 14px;
  box-shadow: var(--card-shadow, 0 1px 3px rgba(0, 0, 0, 0.08));
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
