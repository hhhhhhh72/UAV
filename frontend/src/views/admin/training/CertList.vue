<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索持有人或证书编号..." clearable style="width: 260px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.status" clearable style="width: 140px" @change="onSearchSubmit">
          <el-option label="全部" value="" />
          <el-option label="有效" value="valid" />
          <el-option label="已过期" value="expired" />
          <el-option label="已吊销" value="revoked" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left:auto">
          <el-button type="success" @click="handleAdd">新增</el-button>
        </div>
      </div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <el-table-column prop="cert_number" label="证书编号" width="180">
          <template #default="{ row }">
            <span class="cell-mono">{{ row.cert_no || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="user_id" label="持有人ID" min-width="140" />
        <el-table-column prop="cert_type" label="证书类型" width="120">
          <template #default="{ row }">{{ row.cert_type || '-' }}</template>
        </el-table-column>
        <el-table-column prop="issue_date" label="签发日期" width="120" sortable="custom">
          <template #default="{ row }">{{ formatDate(row.issue_date) }}</template>
        </el-table-column>
        <el-table-column prop="expire_date" label="有效期至" width="120">
          <template #default="{ row }">{{ formatDate(row.expire_date) }}</template>
        </el-table-column>
        <el-table-column prop="level" label="等级" width="80">
          <template #default="{ row }">{{ row.level || '-' }}</template>
        </el-table-column>
        <el-table-column prop="issuer_org" label="发证机构" min-width="140">
          <template #default="{ row }">{{ row.issuer_org || '-' }}</template>
        </el-table-column>
        <el-table-column label="证书图片" width="80">
          <template #default="{ row }">
            <el-image
              v-if="row.image_url"
              :src="fullUrl(row.image_url)"
              :preview-src-list="[fullUrl(row.image_url)]"
              fit="cover"
              style="width: 44px; height: 44px; border-radius: 4px; cursor: pointer;"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel[row.status] || row.status || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无证书数据" /></template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>

    <el-dialog v-model="detailVisible" title="证书详情" width="600px" :close-on-click-modal="false">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="证书编号">{{ currentItem.cert_no || '-' }}</el-descriptions-item>
          <el-descriptions-item label="证书类型">{{ currentItem.cert_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="持有人">{{ currentItem.holder_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="签发日期">{{ formatDate(currentItem.issue_date) }}</el-descriptions-item>
          <el-descriptions-item label="有效期至">{{ formatDate(currentItem.expire_date) }}</el-descriptions-item>
          <el-descriptions-item label="发证机构" :span="2">{{ currentItem.issuer || '-' }}</el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ currentItem.remark || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>

    <el-dialog v-model="formVisible" :title="formEdit?'编辑证书':'新增证书'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="证书编号" required><el-input v-model="form.cert_number" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="证书类型"><el-input v-model="form.cert_type" placeholder="caac / utc_dji / gov_level" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="发证机构"><el-input v-model="form.issuer_org" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="签发日期"><el-date-picker v-model="form.issue_date" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width:100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="有效期至"><el-date-picker v-model="form.expire_date" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="状态"><el-select v-model="form.status"><el-option label="待审核" value="pending" /><el-option label="已通过" value="approved" /><el-option label="已驳回" value="rejected" /></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="formVisible=false">取消</el-button><el-button type="primary" @click="submitForm" :loading="formLoading">提交</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('certificates')

// 相对路径图片补全后端地址（vite/nginx 已代理 /uploads）
const fullUrl = (u) => (u && u.startsWith('http') ? u : (import.meta.env.VITE_API_TARGET || 'http://localhost:8080') + (u || ''))

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ 'valid': 'success', 'expired': 'warning', 'revoked': 'danger' }[s] || 'info')
const statusLabel = { 'valid': '有效', 'expired': '已过期', 'revoked': '已吊销' }

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSelectChange, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { status: '' }
})

const detailVisible = ref(false)
const currentItem = ref(null)

const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible=ref(false); const formEdit=ref(false); const formLoading=ref(false)
const form=reactive({id:'',cert_number:'',cert_type:'',issuer_org:'',issue_date:'',expire_date:'',status:'pending'})
const resetForm=()=>Object.assign(form,{id:'',cert_no:'',holder_name:'',cert_type:'',issuer:'',issue_date:'',expire_date:'',status:'valid',remark:''})
const handleAdd=()=>{resetForm();formEdit.value=false;formVisible.value=true}
const handleEdit=(r)=>{Object.assign(form,r);formEdit.value=true;formVisible.value=true}
const submitForm=async()=>{if(!form.cert_number){ElMessage.warning('请输入证书编号');return};formLoading.value=true;try{const p={...form};formEdit.value?await api.update(form.id,p):await api.create(p);ElMessage.success(formEdit.value?'更新成功':'创建成功');formVisible.value=false;loadData()}catch(e){ElMessage.error(e?.response?.data?.message||'操作失败')}finally{formLoading.value=false}}
const handleDelete=(r)=>{ElMessageBox.confirm('确定删除该证书?','提示',{type:'warning'}).then(async()=>{try{await api.delete(r.id);ElMessage.success('已删除');loadData()}catch{ElMessage.error('删除失败')}}).catch(()=>{})}

onMounted(loadData)
</script>

<style scoped>
.list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.cell-mono { font-family: 'Courier New', monospace; font-size: 13px; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
