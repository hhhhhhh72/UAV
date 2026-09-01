<template>
  <div class="rich-editor">
    <div class="rich-toolbar">
      <button type="button" title="加粗" @mousedown.prevent @click="cmd('bold')"><b>B</b></button>
      <button type="button" title="斜体" @mousedown.prevent @click="cmd('italic')"><i>I</i></button>
      <button type="button" title="下划线" @mousedown.prevent @click="cmd('underline')"><u>U</u></button>
      <button type="button" title="无序列表" @mousedown.prevent @click="cmd('insertUnorderedList')">•</button>
      <button type="button" title="有序列表" @mousedown.prevent @click="cmd('insertOrderedList')">1.</button>
      <button type="button" title="段落" @mousedown.prevent @click="setBlock('p')">P</button>
      <button type="button" title="标题" @mousedown.prevent @click="setBlock('h3')">H</button>
      <button type="button" title="链接" @mousedown.prevent @click="addLink">🔗</button>
      <button type="button" title="清除格式" @mousedown.prevent @click="cmd('removeFormat')">✕</button>
    </div>
    <div
      ref="editorRef"
      class="rich-body"
      contenteditable="true"
      :placeholder="placeholder"
      @input="onInput"
      @keydown.enter.prevent="onEnter"
    ></div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '请输入内容…' },
})
const emit = defineEmits(['update:modelValue'])

const editorRef = ref(null)

const cmd = (name, value) => {
  editorRef.value?.focus()
  document.execCommand(name, false, value || null)
  emit('update:modelValue', editorRef.value.innerHTML)
}
const setBlock = (tag) => cmd('formatBlock', tag)
const addLink = () => {
  const url = prompt('输入链接地址（http/https 或 /开头）')
  if (!url) return
  if (!/^(https?:\/\/|\/|\.\/|#)/i.test(url)) return
  cmd('createLink', url)
}
const onInput = () => emit('update:modelValue', editorRef.value ? editorRef.value.innerHTML : '')
const onEnter = () => cmd('formatBlock', 'p')

watch(
  () => props.modelValue,
  (v) => {
    if (editorRef.value && editorRef.value.innerHTML !== (v || '')) {
      editorRef.value.innerHTML = v || ''
    }
  }
)
onMounted(() => {
  if (editorRef.value) editorRef.value.innerHTML = props.modelValue || ''
})
</script>

<style scoped>
.rich-editor { border: 1px solid var(--color-border, #e5e6eb); border-radius: 4px; overflow: hidden; }
.rich-toolbar { display: flex; flex-wrap: wrap; gap: 2px; padding: 6px; background: #f7f8fa; border-bottom: 1px solid var(--color-border, #e5e6eb); }
.rich-toolbar button { min-width: 28px; height: 26px; border: none; background: #fff; border-radius: 4px; cursor: pointer; font-size: 13px; color: #4e5969; }
.rich-toolbar button:hover { background: #e5e6eb; }
.rich-body { min-height: 180px; padding: 10px 12px; font-size: 13px; line-height: 1.7; outline: none; }
.rich-body:empty::before { content: attr(placeholder); color: #86909c; }
.rich-body p { margin: 0 0 6px; }
</style>
