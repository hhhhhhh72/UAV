<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索标题..." clearable style="width: 220px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.msg_type" clearable placeholder="消息类型" style="width: 160px" @change="onSearchSubmit">
          <el-option label="全部" value="" />
          <el-option label="系统通知" value="系统通知" />
          <el-option label="活动提醒" value="活动提醒" />
          <el-option label="审核结果" value="审核结果" />
          <el-option label="其他" value="其他" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left:auto">
          <el-button type="warning" @click="handleSend">发送通知</el-button>
          <el-button type="success" @click="handleAdd">新增</el-button>
        </div>
      </div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange" @sort-change="onSortChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column prop="msg_type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="typeColor[row.msg_type] || 'info'">{{ row.msg_type || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="to_user" label="接收者" width="140" />
        <el-table-column prop="created_at" label="发送时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusColor[row.status] || 'info'">{{ statusLabel[row.status] || row.status || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无数据" /></template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>

    <el-dialog v-model="detailVisible" title="通知详情" width="640px" destroy-on-close>
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID" :span="2">{{ currentItem.id }}</el-descriptions-item>
          <el-descriptions-item label="标题" :span="2">{{ currentItem.title }}</el-descriptions-item>
          <el-descriptions-item label="消息类型">
            <el-tag :type="typeColor[currentItem.msg_type] || 'info'" size="small">{{ currentItem.msg_type || '-' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="接收者">{{ currentItem.to_user || '-' }}</el-descriptions-item>
          <el-descriptions-item label="发送时间">{{ formatDate(currentItem.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="阅读时间">{{ formatDate(currentItem.read_at) }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusColor[currentItem.status] || 'info'" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="内容" :span="2">
            <div style="white-space: pre-wrap; line-height: 1.6;">{{ currentItem.content || '-' }}</div>
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
    <el-dialog v-model="formVisible" :title="formEdit?'编辑通知':'发送通知'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="90px">
        <el-row :gutter="16"><el-col :span="16"><el-form-item label="消息标题" required><el-input v-model="form.title"/></el-form-item></el-col><el-col :span="8"><el-form-item label="类型"><el-select v-model="form.msg_type" style="width:100%"><el-option label="系统通知" value="系统通知"/><el-option label="活动提醒" value="活动提醒"/><el-option label="审核结果" value="审核结果"/><el-option label="其他" value="其他"/></el-select></el-form-item></el-col></el-row>
        <el-form-item label="接收者"><el-input v-model="form.to_user" placeholder="留空表示全部用户"/></el-form-item>
        <el-form-item label="消息内容" required><el-input v-model="form.content" type="textarea" rows="5"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="formVisible=false">取消</el-button><el-button type="primary" @click="submitForm" :loading="formLoading">发送</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('messages')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  if (isNaN(dt.getTime())) return '-'
  const p = n => String(n).padStart(2, '0')
  return dt.getFullYear() + '-' + p(dt.getMonth() + 1) + '-' + p(dt.getDate()) + ' ' + p(dt.getHours()) + ':' + p(dt.getMinutes())
}

const typeColor = { '系统通知': 'warning', '活动提醒': 'success', '审核结果': 'danger', '其他': 'info' }
const statusLabel = { 'unread': '未读', 'read': '已读' }
const statusColor = { 'unread': 'warning', 'read': 'info' }

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, onSelectChange, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { msg_type: '' }
})

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }
const formVisible=ref(false);const formEdit=ref(false);const formLoading=ref(false)
const form=reactive({id:'',title:'',msg_type:'系统通知',to_user:'',content:'',status:'unread'})
const resetForm=()=>Object.assign(form,{id:'',title:'',msg_type:'系统通知',to_user:'',content:'',status:'unread'})
const handleSend=()=>{resetForm();formEdit.value=false;formVisible.value=true}
const handleAdd=()=>{resetForm();formEdit.value=false;formVisible.value=true}
const handleEdit=(r)=>{Object.assign(form,r);formEdit.value=true;formVisible.value=true}
const submitForm=async()=>{if(!form.title){ElMessage.warning('请输入消息标题');return};formLoading.value=true;try{const p={...form};formEdit.value?await api.update(form.id,p):await api.create(p);ElMessage.success(formEdit.value?'更新成功':'创建成功');formVisible.value=false;loadData()}catch(e){ElMessage.error(e?.response?.data?.message||'操作失败')}finally{formLoading.value=false}}
const handleDelete = (row) => {
  ElMessageBox.confirm('确定删除该通知吗？', '提示', { type: 'warning' }).then(async () => {
    try { await api.delete(row.id); ElMessage.success('已删除'); loadData() } catch { ElMessage.error('删除失败') }
  }).catch(() => {})
}

onMounted(loadData)
</script>

<style scoped>
.list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
