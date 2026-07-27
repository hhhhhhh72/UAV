import { ref, reactive, computed, unref } from 'vue'
import { showFailToast } from 'vant'

/**
 * 通用的列表数据请求 Hook
 * 封装了分页、搜索、排序、选中、批量操作、加载状态
 *
 * 参考 Tigshop 的 useListRequest 模式
 *
 * @param {Object} options
 * @param {Function|ComputedRef<Function>} options.apiFunction - API 请求函数 (params) => Promise<{ data, total }>
 * @param {string} options.idKey - 行唯一标识字段，默认 'id'
 * @param {Object} options.defaultParams - 默认查询参数
 * @param {number} options.defaultPageSize - 默认每页条数，默认 20
 * @returns {Object} { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, onSelectChange, onBatchAction, resetParams }
 *
 * @example
 * const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
 *   apiFunction: getEnterpriseList,
 *   idKey: 'id',
 *   defaultParams: { status: 'submitted' }
 * })
 */
export function useListRequest(options) {
  const { apiFunction, idKey = 'id', defaultParams = {}, defaultPageSize = 20 } = options

  // 保存初始参数，用于重置
  const initialParams = {
    page: 1,
    page_size: defaultPageSize,
    ...defaultParams
  }

  // 响应式状态
  const listData = ref([])
  const loading = ref(false)
  const total = ref(0)
  const selectedIds = ref([])
  const filterParams = reactive({ ...initialParams })

  // 处理动态 API 函数（支持 ComputedRef）
  const resolvedApi = computed(() => {
    const fn = unref(apiFunction)
    return typeof fn === 'function' ? fn : () => Promise.reject(new Error('API function is invalid'))
  })

  /**
   * 加载数据
   */
  const loadData = async () => {
    loading.value = true
    try {
      // Trim keyword
      if (filterParams.keyword) {
        filterParams.keyword = filterParams.keyword.trim()
      }
      const result = await resolvedApi.value(filterParams)

      // 兼容多种后端响应格式
      if (Array.isArray(result)) {
        listData.value = result
        total.value = result.length
      } else if (result && result.data) {
        listData.value = result.data
        total.value = result.total || result.data.length
      } else {
        listData.value = []
        total.value = 0
      }
    } catch (error) {
      listData.value = []
      total.value = 0
      showFailToast(error?.response?.data?.message || error?.message || '请求失败')
    } finally {
      loading.value = false
    }
  }

  /**
   * 搜索提交：重置到第一页
   */
  const onSearchSubmit = () => {
    filterParams.page = 1
    loadData()
  }

  /**
   * 排序变化
   */
  const onSortChange = ({ prop, order }) => {
    filterParams.sort_field = prop
    filterParams.sort_order = order === 'ascending' ? 'asc' : order === 'descending' ? 'desc' : ''
    loadData()
  }

  /**
   * 页码变化
   */
  const onPageChange = (page, size) => {
    filterParams.page = page
    filterParams.page_size = size
    loadData()
  }

  /**
   * 选中行变化
   */
  const onSelectChange = (items) => {
    selectedIds.value = items.map((item) => item[idKey])
  }

  /**
   * 批量操作
   */
  const onBatchAction = async (action, batchApi) => {
    if (selectedIds.value.length === 0) {
      showFailToast('请至少选择一项')
      return
    }
    try {
      await batchApi(action, { ids: selectedIds.value })
      loadData()
      selectedIds.value = []
    } catch (error) {
      showFailToast(error?.response?.data?.message || error?.message || '批量操作失败')
    }
  }

  /**
   * 重置参数到初始值
   */
  const resetParams = () => {
    Object.assign(filterParams, initialParams)
    loadData()
  }

  return {
    listData,
    loading,
    total,
    selectedIds,
    filterParams,
    loadData,
    onSearchSubmit,
    onSortChange,
    onPageChange,
    onSelectChange,
    onBatchAction,
    resetParams
  }
}
