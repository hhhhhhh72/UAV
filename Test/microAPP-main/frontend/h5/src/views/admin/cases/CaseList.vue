<template>
  <div class="case-list-page">
    <DataToolbar>
      <template #filters>
        <span class="toolbar-label">案例列表</span>
      </template>
      <template #actions>
        <van-button type="primary" size="small" icon="plus" @click="createCase">新增案例</van-button>
        <van-button type="warning" size="small" icon="label-o" @click="openCategoryManager">管理分类</van-button>
        <van-button type="default" size="small" icon="replay" @click="refreshAll">刷新</van-button>
      </template>
    </DataToolbar>

    <van-empty v-if="cases.length === 0" description="暂无案例数据" />

    <van-cell-group v-else inset style="border-radius: var(--card-radius);">
      <van-cell v-for="caseItem in cases" :key="caseItem.id" is-link @click="editCase(caseItem)">
        <template #title>
          <div style="display:flex; align-items:center; gap:10px;">
            <div class="case-thumb">
              <img v-if="caseItem.coverType !== 'video'" :src="caseItem.cover" :alt="caseItem.title" />
              <video v-else :src="caseItem.cover" muted playsinline preload="metadata" />
            </div>
            <div style="min-width: 0;">
              <div class="case-title">
                {{ caseItem.title || '未命名案例' }}
                <van-tag v-if="caseItem.subTag" type="warning" size="medium" style="margin-left:6px;">{{ caseItem.subTag }}</van-tag>
              </div>
              <div class="case-meta">分类：{{ getCategoryName(caseItem.categoryId) }} · {{ caseItem.location || '-' }} · {{ caseItem.date || '-' }}</div>
            </div>
          </div>
        </template>
        <template #value>
          <div style="display:flex; flex-direction: column; align-items: flex-end; gap: 6px;">
            <van-tag type="primary" size="medium">{{ caseItem.coverType === 'video' ? '视频' : '图片' }}</van-tag>
            <span style="font-size: 12px; color: var(--text-secondary);">{{ caseItem.views || '-' }}</span>
          </div>
        </template>
      </van-cell>
    </van-cell-group>

    <!-- Case Edit Popup -->
    <van-popup :show="showCaseEditPopup" @update:show="v => showCaseEditPopup = v" position="bottom" :style="{ height: '90%' }" round>
      <div class="detail-content" v-if="currentCase">
        <van-form @submit="onSaveCase">
          <van-cell-group title="基本信息">
            <van-field name="categoryId" label="所属分类" required>
              <template #input>
                <van-radio-group v-model="currentCase.categoryId" direction="horizontal" @change="onCategoryChange">
                  <van-radio v-for="cat in caseCategories" :key="cat.id" :name="cat.id">{{ cat.name }}</van-radio>
                </van-radio-group>
              </template>
            </van-field>
            <van-field v-model="currentCase.title" label="标题" placeholder="请输入标题" required />
            <van-field v-model="currentCase.subTag" label="子标签" placeholder="可选，用于区分同分类下的细分场景" />
            <van-field v-model="currentCase.service" label="服务类型" placeholder="如：无人机物流服务" />
            <van-field v-model="currentCase.description" label="简介" type="textarea" rows="2" placeholder="请输入简介" />
            <van-field v-model="currentCase.location" label="地点" placeholder="请输入地点" />
            <van-field v-model="currentCase.date" label="时间" placeholder="请输入时间" />
          </van-cell-group>

          <van-cell-group title="封面设置">
            <van-field name="coverType" label="封面类型">
              <template #input>
                <van-radio-group v-model="currentCase.coverType" direction="horizontal">
                  <van-radio name="image">图片</van-radio>
                  <van-radio name="video">视频</van-radio>
                </van-radio-group>
              </template>
            </van-field>
            <van-field v-model="currentCase.cover" label="封面地址" placeholder="输入URL或上传" />
            <van-field label="上传封面">
              <template #input>
                <van-uploader :after-read="onReadCover" max-count="1" :accept="currentCase.coverType === 'video' ? 'video/*' : 'image/*'">
                  <van-button icon="plus" type="primary" size="small" plain>上传{{ currentCase.coverType === 'video' ? '视频' : '图片' }}</van-button>
                </van-uploader>
              </template>
            </van-field>
          </van-cell-group>

          <van-cell-group title="媒体资源">
            <div v-for="(media, index) in currentCase.media" :key="index" style="padding: 10px; border-bottom: 1px dashed #eee;">
              <div style="display: flex; justify-content: space-between; margin-bottom: 8px;">
                <span style="font-weight: bold;">资源 #{{ index + 1 }}</span>
                <van-button size="mini" type="danger" icon="cross" @click="currentCase.media.splice(index, 1)" />
              </div>
              <van-radio-group v-model="media.type" direction="horizontal" style="margin-bottom: 8px;">
                <van-radio name="image">图片</van-radio>
                <van-radio name="video">视频</van-radio>
              </van-radio-group>
              <van-field v-model="media.url" label="地址" placeholder="URL" />
              <div style="margin-top: 8px;">
                <van-uploader :after-read="file => onReadMedia(file, index)" :accept="media.type === 'video' ? 'video/*' : 'image/*'">
                  <van-button icon="plus" size="mini" type="default">上传{{ media.type === 'video' ? '视频' : '图片' }}</van-button>
                </van-uploader>
              </div>
            </div>
            <div style="padding: 10px;">
              <van-button size="small" type="primary" block plain icon="plus" @click="currentCase.media.push({ type: 'image', url: '' })">添加资源</van-button>
            </div>
          </van-cell-group>

          <van-cell-group title="项目亮点">
            <div v-for="(tag, index) in currentCase.highlights" :key="index" style="display: flex; align-items: center; padding: 0 16px;">
              <van-field v-model="currentCase.highlights[index]" :label="'标签 ' + (index + 1)" placeholder="输入标签内容" />
              <van-button size="mini" type="danger" icon="cross" @click="currentCase.highlights.splice(index, 1)" style="margin-left: 8px;" />
            </div>
            <div style="padding: 10px;">
              <van-button size="small" type="primary" block plain icon="plus" @click="currentCase.highlights.push('')">添加标签</van-button>
            </div>
          </van-cell-group>

          <van-cell-group title="详细内容">
            <van-field v-model="currentCase.fullDescription" label="详细描述" type="textarea" rows="4" placeholder="请输入详细描述" autosize />
          </van-cell-group>

          <div style="margin: 16px; padding-bottom: 30px;">
            <van-button round block type="primary" native-type="submit" style="margin-bottom: 12px;">保存修改</van-button>
            <van-button v-if="currentCase.id" round block type="danger" native-type="button" @click="onDeleteCase">删除案例</van-button>
          </div>
        </van-form>
      </div>
    </van-popup>

    <!-- Category Manager Popup -->
    <van-popup :show="showCategoryPopup" @update:show="v => showCategoryPopup = v" position="bottom" :style="{ height: '70%' }" round>
      <div class="category-manager">
        <div class="category-header">
          <span class="category-title">分类管理（共 {{ caseCategories.length }} 项）</span>
          <van-button type="primary" size="small" icon="plus" @click="openAddCategoryDialog">新增分类</van-button>
        </div>

        <van-empty v-if="caseCategories.length === 0" description="暂无分类，请点击右上角新增" />

        <div v-else class="category-list">
          <div v-for="(cat, idx) in caseCategories" :key="cat.id" class="category-row">
            <div class="category-row-head">
              <span class="category-row-index">#{{ idx + 1 }}</span>
              <van-tag type="primary" size="medium">ID: {{ cat.id }}</van-tag>
              <div class="category-row-actions">
                <van-button size="mini" type="success" icon="edit" @click="startEditCategory(cat)">编辑</van-button>
                <van-button size="mini" type="danger" icon="delete-o" @click="deleteCategory(cat)">删除</van-button>
              </div>
            </div>
            <van-cell title="分类名称" :value="cat.name || '-'" />
            <van-cell title="默认服务" :value="cat.service || '-'" />
          </div>
        </div>
      </div>
    </van-popup>

    <!-- 新增/编辑分类对话框 -->
    <van-dialog
      :show="showCategoryDialog"
      @update:show="v => showCategoryDialog = v"
      :title="editingCategory && editingCategory.id ? '编辑分类' : '新增分类'"
      show-cancel-button
      :before-close="onCategoryDialogClose"
    >
      <div style="padding: 16px;" v-if="editingCategory">
        <van-field v-model="editingCategory.name" label="分类名称" placeholder="如：共享无人机" required />
        <van-field v-model="editingCategory.service" label="默认服务" placeholder="如：共享无人机服务" />
      </div>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/http'
import { showFailToast, showSuccessToast, showLoadingToast, closeToast, showConfirmDialog } from 'vant'
import DataToolbar from '../components/DataToolbar.vue'
import { normalizeMediaUrl, uploadFile } from '../composables/useMedia'

const cases = ref([])
const showCaseEditPopup = ref(false)
const currentCase = ref(null)

// 案例分类：从后端 /api/case-categories 动态加载
const caseCategories = ref([])

// 分类管理弹窗
const showCategoryPopup = ref(false)
const showCategoryDialog = ref(false)
const editingCategory = ref(null)  // 当前在对话框中编辑的分类（新增时 id=null）

const fetchCaseCategories = async () => {
  try {
    const res = await axios.get('/api/case-categories')
    const list = Array.isArray(res.data) ? res.data : []
    caseCategories.value = list
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
  if (cat) {
    const defaults = caseCategories.value.map(c => c.service).filter(Boolean)
    if (!currentCase.value.service || defaults.includes(currentCase.value.service)) {
      currentCase.value.service = cat.service || ''
    }
  }
}

// ========== 分类管理 ==========
const openCategoryManager = async () => {
  await fetchCaseCategories()
  showCategoryPopup.value = true
}

// 新增：打开对话框（空白表单）
const openAddCategoryDialog = () => {
  editingCategory.value = { id: null, name: '', service: '' }
  showCategoryDialog.value = true
}

// 编辑：打开对话框（带回填）
const startEditCategory = (cat) => {
  editingCategory.value = { id: cat.id, name: cat.name || '', service: cat.service || '' }
  showCategoryDialog.value = true
}

// 对话框「确认/取消」统一处理（vant Dialog 的 before-close 钩子）
const onCategoryDialogClose = async (action) => {
  if (action !== 'confirm') return true

  const form = editingCategory.value
  if (!form || !form.name || !String(form.name).trim()) {
    showFailToast('分类名称不能为空')
    return false // 阻止关闭，让用户继续输入
  }

  showLoadingToast({ message: '保存中...', forbidClick: true })
  try {
    if (form.id == null) {
      await axios.post('/api/case-categories/create', {
        name: String(form.name).trim(),
        service: String(form.service || '').trim()
      })
    } else {
      await axios.post('/api/case-categories/update', {
        id: form.id,
        name: String(form.name).trim(),
        service: String(form.service || '').trim()
      })
    }
    closeToast()
    showSuccessToast(form.id == null ? '分类已新增' : '分类已更新')
    await fetchCaseCategories()
    return true
  } catch (error) {
    closeToast()
    console.error('保存分类失败', error)
    const msg = error?.response?.data?.message || '保存失败'
    showFailToast(msg)
    return false
  }
}

// 删除分类
const deleteCategory = (cat) => {
  showConfirmDialog({
    title: '确认删除',
    message: `确定要删除分类「${cat.name}」吗？\n若有案例仍归属该分类将无法删除。`
  })
    .then(async () => {
      showLoadingToast({ message: '删除中...', forbidClick: true })
      try {
        await axios.post('/api/case-categories/delete', { id: cat.id })
        closeToast()
        showSuccessToast('删除成功')
        await fetchCaseCategories()
      } catch (error) {
        closeToast()
        const msg = error?.response?.data?.message || '删除失败'
        showFailToast(msg)
      }
    })
    .catch(() => {})
}

const refreshAll = async () => {
  await Promise.all([fetchCaseCategories(), fetchCases()])
}

const fetchCases = async () => {
  try {
    const res = await axios.get('/api/cases')
    const raw = Array.isArray(res.data) ? res.data : (res.data?.data || res.data?.cases || [])
    cases.value = Array.isArray(raw)
      ? raw.map(c => ({
          ...c,
          cover: normalizeMediaUrl(c.cover),
          media: Array.isArray(c.media) ? c.media.map(m => ({ ...m, url: normalizeMediaUrl(m?.url) })) : c.media
        }))
      : []
  } catch (error) {
    showFailToast('获取案例数据失败')
    console.error(error)
  }
}

const createCase = () => {
  const firstCat = caseCategories.value[0]
  currentCase.value = {
    title: '', description: '', location: '', date: '', fullDescription: '',
    coverType: 'image', cover: '', media: [], highlights: [],
    categoryId: firstCat ? firstCat.id : null,
    service: firstCat ? firstCat.service : '',
    subTag: ''
  }
  showCaseEditPopup.value = true
}

const editCase = (caseItem) => {
  currentCase.value = JSON.parse(JSON.stringify(caseItem))
  if (!currentCase.value.media) currentCase.value.media = []
  if (!currentCase.value.highlights) currentCase.value.highlights = []
  if (!currentCase.value.coverType) currentCase.value.coverType = 'image'
  if (currentCase.value.subTag == null) currentCase.value.subTag = ''
  showCaseEditPopup.value = true
}

const onReadCover = async (file) => {
  showLoadingToast({ message: '上传中...', forbidClick: true })
  const url = await uploadFile(file)
  closeToast()
  if (url) {
    currentCase.value.cover = normalizeMediaUrl(url)
    showSuccessToast('封面已上传')
  }
}

const onReadMedia = async (file, index) => {
  showLoadingToast({ message: '上传中...', forbidClick: true })
  const url = await uploadFile(file)
  closeToast()
  if (url) {
    currentCase.value.media[index].url = normalizeMediaUrl(url)
    showSuccessToast('资源已上传')
  }
}

const onSaveCase = async () => {
  if (!currentCase.value) return
  try {
    if (currentCase.value.id) {
      await axios.post('/api/cases/update', currentCase.value)
      const index = cases.value.findIndex(c => c.id === currentCase.value.id)
      if (index !== -1) cases.value[index] = currentCase.value
    } else {
      const res = await axios.post('/api/cases/create', currentCase.value)
      currentCase.value.id = res.data.id
      cases.value.unshift(currentCase.value)
    }
    showSuccessToast('保存成功')
    showCaseEditPopup.value = false
  } catch (error) {
    showFailToast('保存失败')
    console.error(error)
  }
}

const onDeleteCase = () => {
  showConfirmDialog({ title: '确认删除', message: '确定要删除这个案例吗？删除后无法恢复。' })
    .then(async () => {
      try {
        await axios.post('/api/cases/delete', { id: currentCase.value.id })
        const index = cases.value.findIndex(c => c.id === currentCase.value.id)
        if (index !== -1) cases.value.splice(index, 1)
        showSuccessToast('删除成功')
        showCaseEditPopup.value = false
      } catch (error) {
        showFailToast('删除失败')
        console.error(error)
      }
    })
    .catch(() => {})
}

onMounted(async () => {
  await fetchCaseCategories()
  await fetchCases()
})
</script>

<style scoped>
.toolbar-label { font-size: 14px; font-weight: 500; color: var(--text-color); }
.case-thumb { width: 56px; height: 56px; border-radius: 8px; overflow: hidden; background: #f7f8fa; flex: 0 0 56px; }
.case-thumb img, .case-thumb video { width: 100%; height: 100%; object-fit: cover; display: block; }
.case-title { font-weight: 600; color: var(--text-color); line-height: 1.2; margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.case-meta { font-size: 12px; color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.detail-content { padding: 16px 0; }

/* 分类管理弹窗 */
.category-manager { padding: 16px 0; display: flex; flex-direction: column; height: 100%; box-sizing: border-box; }
.category-header { display: flex; align-items: center; justify-content: space-between; padding: 0 16px 12px; flex-shrink: 0; }
.category-title { font-size: 16px; font-weight: 600; color: var(--text-color); }
.category-list { flex: 1; overflow-y: auto; padding: 0 16px 16px; }
.category-row { border: 1px solid var(--border-color, #ebedf0); border-radius: 8px; margin-bottom: 12px; overflow: hidden; }
.category-row-head { display: flex; align-items: center; gap: 8px; padding: 10px 12px; background: #f7f8fa; }
.category-row-index { font-size: 12px; color: var(--text-secondary); margin-right: 4px; }
.category-row-actions { margin-left: auto; display: flex; gap: 6px; }
</style>
