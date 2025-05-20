<template>
  <div class="data-selector">
    <button
      v-for="option in options"
      :key="option.value"
      :class="['data-button', { active: selectedType === option.value }]"
      @click="selectType(option.value)"
    >
      {{ option.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

export type DataType = 'chlorophyll' | 'v_current' | 'u_current' | 'combined_current';

interface DataTypeOption {
  value: DataType;
  label: string;
}

const options: DataTypeOption[] = [
  { value: 'chlorophyll', label: 'Chlorophyll' },
  { value: 'v_current', label: 'North-South Current' },
  { value: 'u_current', label: 'West-East Current' },
  { value: 'combined_current', label: 'Combined Current' }
];

interface Props {
  modelValue: DataType;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: 'update:modelValue', value: DataType): void;
}>();

const selectedType = ref(props.modelValue);

function selectType(type: DataType) {
  selectedType.value = type;
  emit('update:modelValue', type);
}
</script>

<style scoped>
.data-selector {
  position: absolute;
  top: 10px;
  left: 10px;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.data-button {
  padding: 8px 12px;
  background-color: white;
  border: 1px solid #ccc;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.data-button:hover {
  background-color: #f0f0f0;
}

.data-button.active {
  background-color: #2196F3;
  color: white;
  border-color: #0d8bf2;
}
</style>
