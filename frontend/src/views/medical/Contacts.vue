<template>
  <div class="contacts-page">
    <van-nav-bar title="常用联系人" left-arrow @click-left="$router.back()" fixed placeholder>
      <template #right>
        <van-icon name="plus" size="20" @click="showAddDialog = true" />
      </template>
    </van-nav-bar>

    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <div class="contact-list" v-if="contacts.length">
        <van-swipe-cell v-for="c in contacts" :key="c.id">
          <van-cell :title="c.name" :label="`${c.phone}${c.org_name ? ' | ' + c.org_name : ''}`" @click="onEdit(c)">
            <template #right-icon>
              <van-tag v-if="c.label" plain type="primary" size="small">{{ c.label }}</van-tag>
            </template>
          </van-cell>
          <template #right>
            <van-button square type="danger" text="删除" @click="onDelete(c.id)" />
          </template>
        </van-swipe-cell>
      </div>
      <van-empty v-else description="暂无常用联系人" />
    </van-pull-refresh>

    <!-- 新增/编辑弹窗 -->
    <van-dialog v-model:show="showAddDialog" :title="editingId ? '编辑联系人' : '新增联系人'" show-cancel-button @confirm="onSave" :before-close="beforeClose">
      <van-form ref="dialogForm">
        <van-field v-model="editForm.name" label="姓名" placeholder="请输入姓名" :rules="[{ required: true }]" />
        <van-field v-model="editForm.phone" label="电话" placeholder="请输入电话" type="tel" :rules="[{ required: true }]" />
        <van-field v-model="editForm.org_name" label="机构" placeholder="选填" />
        <van-field v-model="editForm.label" label="标签" placeholder="如：检验科张护士（选填）" />
      </van-form>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { showSuccessToast, showFailToast, showConfirmDialog } from 'vant'
import { useMedicalStore } from '@/stores/medical'

const store = useMedicalStore()
const contacts = computed(() => store.contacts)
const refreshing = ref(false)
const showAddDialog = ref(false)
const editingId = ref(null)
const editForm = reactive({ name: '', phone: '', org_name: '', label: '' })

onMounted(() => store.fetchContacts())

async function onRefresh() {
  await store.fetchContacts()
  refreshing.value = false
}

function onEdit(c) {
  editingId.value = c.id
  Object.assign(editForm, { name: c.name, phone: c.phone, org_name: c.org_name || '', label: c.label || '' })
  showAddDialog.value = true
}

async function onSave() {
  if (!editForm.name || !editForm.phone) {
    showFailToast('姓名和电话必填')
    return
  }
  try {
    if (editingId.value) {
      await store.updateContact(editingId.value, editForm)
      showSuccessToast('已更新')
    } else {
      await store.addContact(editForm)
      showSuccessToast('已添加')
    }
  } catch (e) {
    showFailToast(e.response?.data?.message || '操作失败')
  }
  resetForm()
}

function beforeClose(action) {
  if (action === 'cancel') { resetForm(); return true }
  return true
}

function resetForm() {
  editingId.value = null
  Object.assign(editForm, { name: '', phone: '', org_name: '', label: '' })
}

async function onDelete(id) {
  try {
    await showConfirmDialog({ message: '确定删除该联系人？' })
    await store.deleteContact(id)
    showSuccessToast('已删除')
  } catch (e) { /* cancelled */ }
}
</script>

<style scoped>
.contacts-page {
  min-height: 100vh;
  background: #f5f5f5;
  max-width: var(--page-max-width);
  margin: 0 auto;
}

/* 固定导航栏居中约束 */
.contacts-page :deep(.van-nav-bar--fixed) {
  left: 50% !important;
  transform: translateX(-50%) !important;
  width: 100% !important;
  max-width: var(--page-max-width);
}

.contact-list {
  padding: 12px 0;
}
</style>
