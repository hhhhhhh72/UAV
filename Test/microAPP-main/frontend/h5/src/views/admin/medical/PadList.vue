<template>
  <div class="pad-list">
    <DataToolbar>
      <template #filters>
        <span class="toolbar-label">起降场管理</span>
      </template>
      <template #actions>
        <van-button type="primary" size="small" icon="plus" @click="createPad">新增起降场</van-button>
        <van-button type="default" size="small" icon="replay" @click="fetchPads">刷新</van-button>
      </template>
    </DataToolbar>

    <!-- 统计 -->
    <div class="stats-row">
      <div class="stat-card">
        <span class="stat-value">{{ pads.length }}</span>
        <span class="stat-label">总数</span>
      </div>
      <div class="stat-card enabled">
        <span class="stat-value">{{ pads.filter(p => p.enabled).length }}</span>
        <span class="stat-label">已启用</span>
      </div>
      <div class="stat-card disabled">
        <span class="stat-value">{{ pads.filter(p => !p.enabled).length }}</span>
        <span class="stat-label">已禁用</span>
      </div>
    </div>

    <!-- 列表 -->
    <van-empty v-if="!loading && pads.length === 0" description="暂无起降场数据" />

    <van-cell-group v-for="pad in pads" :key="pad.id" inset style="margin-bottom: 12px; border-radius: 10px;">
      <van-cell is-link @click="editPad(pad)">
        <template #title>
          <div class="pad-header">
            <span class="pad-name">{{ pad.name }}</span>
            <van-tag :type="pad.enabled ? 'success' : 'default'" size="medium">{{ pad.enabled ? '启用' : '禁用' }}</van-tag>
          </div>
        </template>
        <template #label>
          <div class="pad-info">
            <div class="pad-address">{{ pad.address }}</div>
            <div class="pad-meta">
              <span v-if="pad.contact_name">{{ pad.contact_name }} {{ pad.contact_phone }}</span>
              <span v-if="pad.max_weight">限重 {{ pad.max_weight }}kg</span>
              <span v-if="pad.operating_hours">{{ pad.operating_hours }}</span>
            </div>
            <div class="pad-coords" v-if="pad.latitude && pad.longitude">
              坐标: {{ pad.latitude }}, {{ pad.longitude }}
            </div>
          </div>
        </template>
      </van-cell>
    </van-cell-group>

    <!-- 编辑弹窗 -->
    <van-popup v-model:show="editVisible" position="bottom" :style="{ height: '90%' }" round>
      <div class="edit-popup">
        <div class="edit-header">
          <h3>{{ isCreate ? '新增起降场' : '编辑起降场' }}</h3>
        </div>

        <van-form @submit="onSave">
          <van-cell-group title="基本信息" inset>
            <van-field v-model="form.name" label="名称" placeholder="请输入起降场名称" required :rules="[{ required: true, message: '请输入名称' }]" />
            <van-field v-model="form.address" label="地址" placeholder="请输入详细地址" required :rules="[{ required: true, message: '请输入地址' }]" />
            <van-field v-model="form.latitude" label="纬度" placeholder="如: 30.5728" type="number" />
            <van-field v-model="form.longitude" label="经度" placeholder="如: 104.0668" type="number" />
          </van-cell-group>

          <van-cell-group title="联系信息" inset>
            <van-field v-model="form.contact_name" label="联系人" placeholder="请输入联系人姓名" />
            <van-field v-model="form.contact_phone" label="联系电话" placeholder="请输入联系电话" type="tel" />
          </van-cell-group>

          <van-cell-group title="运营参数" inset>
            <van-field v-model="form.operating_hours" label="运营时间" placeholder="如: 08:00-18:00" />
            <van-field v-model="form.max_weight" label="最大载重(kg)" placeholder="如: 5" type="number" />
            <van-field name="enabled" label="启用状态">
              <template #input>
                <van-switch v-model="form.enabled" />
              </template>
            </van-field>
          </van-cell-group>

          <div style="margin: 16px; padding-bottom: 30px;">
            <van-button round block type="primary" native-type="submit" style="margin-bottom: 12px;">
              {{ isCreate ? '创建' : '保存修改' }}
            </van-button>
            <van-button v-if="!isCreate" round block type="danger" native-type="button" @click="onDelete">删除起降场</van-button>
          </div>
        </van-form>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { showSuccessToast, showFailToast, showConfirmDialog } from 'vant'
import axios from '@/utils/http'
import DataToolbar from '../components/DataToolbar.vue'

const pads = ref([])
const loading = ref(false)
const editVisible = ref(false)
const isCreate = ref(false)
const currentPadId = ref(null)

const defaultForm = {
  name: '',
  address: '',
  latitude: '',
  longitude: '',
  contact_name: '',
  contact_phone: '',
  operating_hours: '',
  max_weight: '',
  enabled: true
}

const form = reactive({ ...defaultForm })

const resetForm = () => {
  Object.assign(form, defaultForm)
}

const fetchPads = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/medical/pads/all')
    if (res.data?.success) {
      pads.value = res.data.data || []
    }
  } catch (err) {
    showFailToast('获取起降场列表失败')
  } finally {
    loading.value = false
  }
}

const createPad = () => {
  isCreate.value = true
  currentPadId.value = null
  resetForm()
  editVisible.value = true
}

const editPad = (pad) => {
  isCreate.value = false
  currentPadId.value = pad.id
  Object.assign(form, {
    name: pad.name || '',
    address: pad.address || '',
    latitude: pad.latitude || '',
    longitude: pad.longitude || '',
    contact_name: pad.contact_name || '',
    contact_phone: pad.contact_phone || '',
    operating_hours: pad.operating_hours || '',
    max_weight: pad.max_weight || '',
    enabled: pad.enabled !== false
  })
  editVisible.value = true
}

const onSave = async () => {
  const payload = {
    ...form,
    latitude: form.latitude ? parseFloat(form.latitude) : null,
    longitude: form.longitude ? parseFloat(form.longitude) : null,
    max_weight: form.max_weight ? parseFloat(form.max_weight) : null
  }

  try {
    let res
    if (isCreate.value) {
      res = await axios.post('/api/medical/pads', payload)
    } else {
      res = await axios.put(`/api/medical/pads/${currentPadId.value}`, payload)
    }
    if (res.data?.success) {
      showSuccessToast(isCreate.value ? '创建成功' : '保存成功')
      editVisible.value = false
      fetchPads()
    } else {
      showFailToast(res.data?.message || '操作失败')
    }
  } catch (err) {
    showFailToast('操作失败')
  }
}

const onDelete = async () => {
  try {
    await showConfirmDialog({ title: '确认删除', message: '删除后不可恢复，确认删除该起降场？' })
    const res = await axios.delete(`/api/medical/pads/${currentPadId.value}`)
    if (res.data?.success) {
      showSuccessToast('已删除')
      editVisible.value = false
      fetchPads()
    } else {
      showFailToast(res.data?.message || '删除失败')
    }
  } catch (err) {
    // 用户取消 or 请求失败
    if (err !== 'cancel') {
      showFailToast('删除失败')
    }
  }
}

onMounted(fetchPads)
</script>

<style scoped>
.pad-list {
  padding-bottom: 20px;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}

.stat-card {
  background: #fff;
  border-radius: 10px;
  padding: 14px;
  text-align: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.stat-card .stat-value {
  display: block;
  font-size: 24px;
  font-weight: 700;
  color: var(--accent-color, #0071e3);
}

.stat-card.enabled .stat-value { color: #34c759; }
.stat-card.disabled .stat-value { color: #86868b; }

.stat-label {
  display: block;
  font-size: 12px;
  color: var(--text-secondary, #86868b);
  margin-top: 4px;
}

.pad-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pad-name {
  font-weight: 600;
  font-size: 15px;
}

.pad-info {
  margin-top: 4px;
}

.pad-address {
  font-size: 13px;
  color: var(--text-color, #1d1d1f);
}

.pad-meta {
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: var(--text-secondary, #86868b);
  margin-top: 4px;
}

.pad-coords {
  font-size: 11px;
  color: var(--text-secondary, #86868b);
  margin-top: 2px;
  font-family: monospace;
}

.edit-popup {
  padding: 20px;
  overflow-y: auto;
  height: 100%;
}

.edit-header {
  margin-bottom: 16px;
}

.edit-header h3 {
  margin: 0;
  font-size: 18px;
}
</style>
