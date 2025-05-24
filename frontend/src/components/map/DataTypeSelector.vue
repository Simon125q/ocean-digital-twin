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

    <div class="raw-data-container">
      <div class="raw-data-switch">
        <span>Show raw data</span>
        <label class="switch">
          <input type="checkbox" v-model="showRawData" @change="toggleRawData" />
          <span class="slider round"></span>
        </label>
      </div>
    </div>
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
  showRawData: boolean; // Add showRawData prop
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: 'update:modelValue', value: DataType): void;
  (e: 'update:showRawData', value: boolean): void; // Emit event for raw data toggle
}>();

const selectedType = ref(props.modelValue);
const showRawData = ref(props.showRawData); // Initialize from prop

function selectType(type: DataType) {
  selectedType.value = type;
  emit('update:modelValue', type);
}

function toggleRawData() {
  emit('update:showRawData', showRawData.value);
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

.raw-data-container {
  background-color: white;
  border-radius: 8px;
  padding: 10px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.raw-data-switch {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #333;

}

/* The switch - the box around the slider */
.switch {
  position: relative;
  display: inline-block;
  width: 30px;
  height: 17px;
}

/* Hide default HTML checkbox */
.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

/* The slider */
.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  -webkit-transition: .4s;
  transition: .4s;
}

.slider:before {
  position: absolute;
  content: "";
  height: 13px;
  width: 13px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  -webkit-transition: .4s;
  transition: .4s;
}

input:checked+.slider {
  background-color: #2196F3;
}

input:focus+.slider {
  box-shadow: 0 0 1px #2196F3;
}

input:checked+.slider:before {
  -webkit-transform: translateX(13px);
  -ms-transform: translateX(13px);
  transform: translateX(13px);
}

/* Rounded sliders */
.slider.round {
  border-radius: 17px;
}

.slider.round:before {
  border-radius: 50%;
}
</style>

