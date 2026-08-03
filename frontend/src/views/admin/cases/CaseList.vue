<template>
  <div class="case-list-page">
    <!-- 操作栏 -->
    <div class="search-bar">
      <div class="search-row">
        <el-select v-model="filterParams.categoryId" clearable style="width: 160px" @change="onSearchSubmit">
          <el-option label="全部分类" :value="null" />
          <el-option v-for="cat in caseCategories" :key="cat.id" :label="cat.name" :value="Number(cat.id)" />
        </el-select>

        <el-input
          v-model="filterParams.keyword"
          placeholder="搜索案例标题..."
          clearable style="width: 200px"
          @keyup.enter="onSearchSubmit" @clear="onSearchSubmit"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>

        <div style="margin-left: auto; display: flex; gap: 8px;">
          <el-button type="primary" :icon="Plus" @click="createCase">新增案例</el-button>
          <el-button type="warning" :icon="Operation" @click="openCategoryManager">管理分类</el-button>
          <el-button :icon="RefreshRight" @click="refreshAll">刷新</el-button>
        </div>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border>
        <el-table-column label="封面" width="90">
          <template #default="{ row }">
            <div class="case-thumb">
              <img v-if="row.coverType !== 'video'" :src="row.cover" :alt="row.title" />
              <video v-else :src="row.cover" muted playsinline preload="metadata" />
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="title" label="标题" min-width="180" sortable="custom">
          <template #default="{ row }">
            <span class="cell-title">{{ row.title || '未命名案例' }}</span>
            <el-tag v-if="row.subTag" type="warning" size="small" style="margin-left: 6px;">{{ row.subTag }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="分类" width="110">
          <template #default="{ row }">{{ getCategoryName(row.categoryId) }}</template>
        </el-table-column>

        <el-table-column prop="location" label="地点" width="120" />
        <el-table-column prop="date" label="时间" width="110" />

        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.coverType === 'video' ? 'success' : ''" size="small">
              {{ row.coverType === 'video' ? '视频' : '图片' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="views" label="浏览量" width="90" align="right" sortable="custom" />

        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="caseStatusColor[row.status]" size="small">{{ caseStatusLabel[row.status] || row.status }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="editCase(row)">编辑</el-button>
            <el-divider direction="vertical" />
            <el-button link type="danger" size="small" @click="onDeleteCase(row)">删除</el-button>
          </template>
        </el-table-column>

        <template #empty><el-empty description="暂无案例数据" /></template>
      </el-table>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination
        v-model:current-page="filterParams.page"
        v-model:page-size="filterParams.page_size"
        :page-sizes="[10, 20, 50]"
        :total="total" layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="loadData" @current-change="loadData"
      />
    </div>

    <!-- 案例编辑弹窗 (保留原有逻辑) -->
    <el-dialog
      v-model="showCaseEditPopup"
      :title="currentCase?.id ? '编辑案例' : '新增案例'"
      width="720px"
      :close-on-click-modal="false"
    >
      <template v-if="currentCase">
        <el-form label-width="80px">
          <el-divider content-position="left">基本信息</el-divider>
          <el-form-item label="所属分类" required>
            <el-radio-group v-model="currentCase.categoryId" @change="onCategoryChange">
              <el-radio v-for="cat in caseCategories" :key="cat.id" :value="Number(cat.id)">{{ cat.name }}</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="标题" required>
            <el-input v-model="currentCase.title" placeholder="请输入标题" />
          </el-form-item>
          <el-form-item label="子标签">
            <el-input v-model="currentCase.subTag" placeholder="可选" />
          </el-form-item>
          <el-form-item label="服务类型">
            <el-input v-model="currentCase.service" placeholder="如：无人机物流服务" />
          </el-form-item>
          <el-form-item label="简介">
            <el-input v-model="currentCase.description" type="textarea" :rows="2" placeholder="请输入简介" />
          </el-form-item>
          <el-form-item label="地点">
            <el-input v-model="currentCase.location" placeholder="请输入地点" />
          </el-form-item>
          <el-form-item label="时间">
            <el-date-picker v-model="currentCase.date" type="date" placeholder="选择日期" format="YYYY-MM-DD" value-format="YYYY-MM-DD" style="width:100%" />
          </el-form-item>

          <el-divider content-position="left">封面设置</el-divider>
          <el-form-item label="封面类型">
            <el-radio-group v-model="currentCase.coverType">
              <el-radio value="image">图片</el-radio>
              <el-radio value="video">视频</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="封面地址">
            <template v-if="currentCase.coverType === 'image'">
              <el-upload class="cover-upload" :action="uploadUrl" :headers="uploadHeaders" :show-file-list="false"
                :on-success="onCoverUploadSuccess" :before-upload="beforeUpload" accept="image/*">
                <img v-if="currentCase.cover" :src="currentCase.cover" class="cover-preview" />
                <el-button v-else type="primary" plain>点击上传</el-button>
              </el-upload>
              <el-button v-if="currentCase.cover" size="small" style="margin-top:8px" @click="currentCase.cover=''">清除</el-button>
            </template>
            <el-input v-else v-model="currentCase.cover" placeholder="输入视频URL" />
          </el-form-item>

          <el-divider content-position="left">审核状态</el-divider>
          <el-form-item label="状态">
            <el-select v-model="currentCase.status" style="width:200px">
              <el-option label="待审核" value="pending" />
              <el-option label="已发布" value="published" />
              <el-option label="已下架" value="archived" />
            </el-select>
          </el-form-item>

          <el-divider content-position="left">详细内容</el-divider>
          <el-form-item label="详细描述">
            <el-input v-model="currentCase.fullDescription" type="textarea" :rows="4" placeholder="请输入详细描述" />
          </el-form-item>

          <!-- 亮点标签 -->
          <el-divider content-position="left">项目亮点</el-divider>
          <div v-for="(tag, idx) in currentCase.highlights" :key="idx" style="margin-bottom: 8px; display: flex; gap: 8px;">
            <el-input v-model="currentCase.highlights[idx]" :placeholder="'标签 ' + (idx + 1)" style="flex: 1;" />
            <el-button type="danger" :icon="Delete" circle size="small" @click="currentCase.highlights.splice(idx, 1)" />
          </div>
          <el-button type="primary" :icon="Plus" plain size="small" @click="currentCase.highlights.push('')">添加标签</el-button>

          <!-- 媒体资源 -->
          <el-divider content-position="left">媒体资源</el-divider>
          <div v-for="(media, idx) in currentCase.media" :key="idx" style="margin-bottom: 12px; padding: 10px; border: 1px dashed #e0e0e0; border-radius: 6px;">
            <div style="display: flex; justify-content: space-between; margin-bottom: 8px;">
              <b>资源 #{{ idx + 1 }}</b>
              <el-button type="danger" size="small" :icon="Delete" circle @click="currentCase.media.splice(idx, 1)" />
            </div>
            <el-radio-group v-model="media.type" style="margin-bottom: 8px;">
              <el-radio value="image">图片</el-radio>
              <el-radio value="video">视频</el-radio>
            </el-radio-group>
            <el-input v-model="media.url" placeholder="URL" />
          </div>
          <el-button type="primary" :icon="Plus" plain size="small" @click="currentCase.media.push({ type: 'image', url: '' })">添加资源</el-button>
        </el-form>
      </template>
      <template #footer>
        <el-button @click="showCaseEditPopup = false">取消</el-button>
        <el-button v-if="currentCase?.id" type="danger" @click="onDeleteCase(currentCase)">删除案例</el-button>
        <el-button type="primary" @click="onSaveCase">保存</el-button>
      </template>
    </el-dialog>

    <!-- 分类管理弹窗 (保留原有逻辑) -->
    <el-dialog v-model="showCategoryPopup" title="分类管理" width="500px">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <span>共 {{ caseCategories.length }} 项</span>
        <el-button type="primary" size="small" :icon="Plus" @click="openAddCategoryDialog">新增分类</el-button>
      </div>

      <el-empty v-if="caseCategories.length === 0" description="暂无分类" />

      <div v-else>
        <div v-for="(cat, idx) in caseCategories" :key="cat.id" style="padding: 10px; border: 1px solid #ebeef5; border-radius: 6px; margin-bottom: 8px;">
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="color: #909399; font-size: 12px;">#{{ idx + 1 }}</span>
            <el-tag type="primary" size="small">ID: {{ cat.id }}</el-tag>
            <span style="flex: 1;">{{ cat.name || '-' }}</span>
            <span style="color: #909399; font-size: 12px;">{{ cat.service || '-' }}</span>
            <el-button size="small" type="success" @click="startEditCategory(cat)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteCategory(cat)">删除</el-button>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 分类编辑子弹窗 -->
    <el-dialog v-model="showCategoryDialog" :title="editingCategory?.id ? '编辑分类' : '新增分类'" width="400px">
      <el-form v-if="editingCategory" label-width="80px">
        <el-form-item label="分类名称" required>
          <el-input v-model="editingCategory.name" placeholder="如：共享无人机" />
        </el-form-item>
        <el-form-item label="默认服务">
          <el-input v-model="editingCategory.service" placeholder="如：共享无人机服务" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCategoryDialog = false">取消</el-button>
        <el-button type="primary" @click="onSaveCategory">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/http'
import { Search, Plus, Operation, RefreshRight, Delete } from '@element-plus/icons-vue'
import { showFailToast, showSuccessToast, showLoadingToast, closeToast, showConfirmDialog } from '@/utils/feedback'
import { useListRequest } from '@/hooks/useListRequest'
import { normalizeMediaUrl } from '../composables/useMedia'

// --- 分类数据 ---
const caseCategories = ref([])

const fetchCaseCategories = async () => {
  try {
    const res = await axios.get('/api/case-categories')
    caseCategories.value = Array.isArray(res.data) ? res.data : []
  } catch (error) {
    console.error('获取分类失败', error)
    caseCategories.value = []
  }
}

const getCategoryName = (id) => {
  const cat = caseCategories.value.find(c => Number(c.id) === Number(id))
  return cat ? cat.name : (id ?? '-')
}

const onCategoryChange = (val) => {
  if (!currentCase.value) return
  const cat = caseCategories.value.find(c => Number(c.id) === Number(val))
  if (cat && !currentCase.value.service) {
    currentCase.value.service = cat.service || ''
  }
}

// --- 案例列表 ---
const fetchCases = async (params) => {
  try {
    const res = await axios.get('/api/cases', { params })
    const raw = res.data?.data || res.data || []
    return {
      data: Array.isArray(raw) ? raw.map(c => ({
        ...c,
        cover: normalizeMediaUrl(c.cover),
        media: Array.isArray(c.media) ? c.media.map(m => ({ ...m, url: normalizeMediaUrl(m?.url) })) : c.media
      })) : [],
      total: res.data?.total || (Array.isArray(raw) ? raw.length : 0)
    }
  } catch (error) {
    showFailToast('获取案例数据失败')
    return { data: [], total: 0 }
  }
}

// 前端搜索过滤包装层
const caseListData = ref([])

const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: computed(() => fetchCases),
  idKey: 'id',
  defaultParams: { categoryId: '' }
})

const refreshAll = async () => {
  await fetchCaseCategories()
  loadData()
}

// --- 案例编辑 ---
const showCaseEditPopup = ref(false)
const currentCase = ref(null)
const uploadUrl = '/api/v1/upload'
const uploadHeaders = { Authorization: `Bearer ${localStorage.getItem('accessToken') || ''}` }
const caseStatusColor = { pending: 'warning', published: 'success', archived: 'info' }
const caseStatusLabel = { pending: '待审核', published: '已发布', archived: '已下架' }

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) { showFailToast('只能上传图片文件'); return false }
  if (!isLt5M) { showFailToast('图片不能超过 5MB'); return false }
  return true
}

const onCoverUploadSuccess = (res) => {
  if (currentCase.value) currentCase.value.cover = res?.data?.url || res?.url || ''
  showSuccessToast('上传成功')
}

const createCase = async () => {
  if (caseCategories.value.length === 0) await fetchCaseCategories()
  const firstCat = caseCategories.value[0]
  currentCase.value = {
    title: '', description: '', location: '', date: '', fullDescription: '',
    coverType: 'image', cover: '', media: [], highlights: [], status: 'pending',
    categoryId: firstCat ? Number(firstCat.id) : null,
    service: firstCat ? firstCat.service : '', subTag: ''
  }
  showCaseEditPopup.value = true
}

const editCase = async (caseItem) => {
  if (caseCategories.value.length === 0) await fetchCaseCategories()
  currentCase.value = JSON.parse(JSON.stringify(caseItem))
  if (!currentCase.value.media) currentCase.value.media = []
  if (!currentCase.value.highlights) currentCase.value.highlights = []
  if (!currentCase.value.coverType) currentCase.value.coverType = 'image'
  if (currentCase.value.subTag == null) currentCase.value.subTag = ''
  showCaseEditPopup.value = true
}

const onSaveCase = async () => {
  if (!currentCase.value) return
  showLoadingToast({ message: '保存中...', forbidClick: true })
  try {
    if (currentCase.value.id) {
      await axios.post('/api/cases/update', currentCase.value)
    } else {
      await axios.post('/api/cases/create', currentCase.value)
    }
    closeToast()
    showSuccessToast('保存成功')
    showCaseEditPopup.value = false
    loadData()
  } catch (error) {
    closeToast()
    showFailToast(error?.response?.data?.message || '保存失败')
  }
}

const onDeleteCase = (caseItem) => {
  showConfirmDialog({ title: '确认删除', message: '确定要删除这个案例吗？删除后无法恢复。' })
    .then(async () => {
      try {
        await axios.post('/api/cases/delete', { id: caseItem.id })
        showSuccessToast('删除成功')
        if (currentCase.value?.id === caseItem.id) showCaseEditPopup.value = false
        loadData()
      } catch (error) {
        showFailToast('删除失败')
      }
    })
    .catch(() => {})
}

// --- 分类管理 ---
const showCategoryPopup = ref(false)
const showCategoryDialog = ref(false)
const editingCategory = ref(null)

const openCategoryManager = async () => {
  await fetchCaseCategories()
  showCategoryPopup.value = true
}

const openAddCategoryDialog = () => {
  editingCategory.value = { id: null, name: '', service: '' }
  showCategoryDialog.value = true
}

const startEditCategory = (cat) => {
  editingCategory.value = { id: cat.id, name: cat.name || '', service: cat.service || '' }
  showCategoryDialog.value = true
}

const onSaveCategory = async () => {
  const form = editingCategory.value
  if (!form || !form.name?.trim()) {
    showFailToast('分类名称不能为空')
    return
  }
  showLoadingToast({ message: '保存中...', forbidClick: true })
  try {
    if (form.id == null) {
      await axios.post('/api/case-categories/create', { name: form.name.trim(), service: (form.service || '').trim() })
    } else {
      await axios.post('/api/case-categories/update', { id: form.id, name: form.name.trim(), service: (form.service || '').trim() })
    }
    closeToast()
    showSuccessToast(form.id == null ? '分类已新增' : '分类已更新')
    showCategoryDialog.value = false
    await fetchCaseCategories()
  } catch (error) {
    closeToast()
    showFailToast(error?.response?.data?.message || '保存失败')
  }
}

const deleteCategory = (cat) => {
  showConfirmDialog({ title: '确认删除', message: `确定要删除分类「${cat.name}」吗？若有案例仍归属该分类将无法删除。` })
    .then(async () => {
      showLoadingToast({ message: '删除中...', forbidClick: true })
      try {
        await axios.post('/api/case-categories/delete', { id: cat.id })
        closeToast()
        showSuccessToast('删除成功')
        await fetchCaseCategories()
      } catch (error) {
        closeToast()
        showFailToast(error?.response?.data?.message || '删除失败')
      }
    })
    .catch(() => {})
}

onMounted(refreshAll)
</script>

<style scoped>
.case-list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.case-thumb { width: 56px; height: 56px; border-radius: 6px; overflow: hidden; background: #f7f8fa; }
.case-thumb img, .case-thumb video { width: 100%; height: 100%; object-fit: cover; display: block; }
.cell-title { font-weight: 500; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.cover-upload { display: inline-block; }
.cover-preview { width: 160px; height: 100px; object-fit: cover; border: 1px solid #ddd; border-radius: 4px; cursor: pointer; }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
