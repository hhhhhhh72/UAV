<template>
  <view class="u-tab" @click="onTap"><slot /></view>
</template>

<script setup>
import { defineEmits } from 'vue'
defineProps({ title: { type: String, default: '' } })
const emit = defineEmits(['tap'])
function onTap() { emit('tap') }
</script>
