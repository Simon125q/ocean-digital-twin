<template>
  <div class="legend-container" :style="legendStyle">
    <div v-if="title" class="legend-title">{{ title }}</div>
    <div class="legend-body">
      <div class="color-bar-wrapper">
        <div class="color-bar" :style="gradientStyle"></div>
      </div>
      <div class="labels-container">
        <div
          v-for="item in processedColorStops"
          :key="item.value + '-' + item.originalIndex"
          class="label-item"
          :style="{ top: item.positionPercent + '%' }"
        >
          <span class="label-text">{{ item.value }} {{ unit }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  colorStops: {
    type: Array,
    required: true,
    validator: (stops) => {
      return (
        Array.isArray(stops) &&
        stops.length > 0 &&
        stops.every(
          (stop) =>
            Array.isArray(stop) &&
            stop.length === 2 &&
            typeof stop[0] === 'number' &&
            typeof stop[1] === 'string'
        )
      );
    },
  },
  unit: {
    type: String,
    default: 'm/s',
  },
  title: {
    type: String,
    default: '',
  },
  width: {
    type: String,
    default: '150px',
  },
  height: {
    type: String,
    default: '200px',
  },
  barWidth: {
    type: String,
    default: '20px',
  },
});

const legendStyle = computed(() => ({
  width: props.width,
}));

// Still sort by value to maintain a logical order for colors and labels if input isn't sorted.
// Add originalIndex to ensure unique keys if values are identical after sorting.
const sortedColorStops = computed(() => {
  return [...props.colorStops]
    .map((stop, index) => ({ value: stop[0], color: stop[1], originalIndex: index }))
    .sort((a, b) => a.value - b.value);
});

const gradientStyle = computed(() => {
  const numStops = sortedColorStops.value.length;

  if (numStops === 0) {
    return { background: 'transparent', height: props.height, width: props.barWidth };
  }
  if (numStops === 1) {
    return {
      background: sortedColorStops.value[0].color,
      height: props.height,
      width: props.barWidth,
    };
  }

  const gradientParts = sortedColorStops.value
    .map((stop, index) => {
      const percentage = (index / (numStops - 1)) * 100;
      return `${stop.color} ${percentage.toFixed(2)}%`;
    })
    .join(', ');

  return {
    background: `linear-gradient(to top, ${gradientParts})`,
    height: props.height,
    width: props.barWidth,
  };
});

const processedColorStops = computed(() => {
  const numStops = sortedColorStops.value.length;

  if (numStops === 0) return [];

  if (numStops === 1) {
    return [
      {
        value: sortedColorStops.value[0].value,
        color: sortedColorStops.value[0].color,
        positionPercent: 50, // Center the single label
        originalIndex: sortedColorStops.value[0].originalIndex,
      },
    ];
  }

  return sortedColorStops.value.map((stop, index) => {
    // Calculate position from the bottom, then convert to 'top' CSS value
    const percentageFromBottom = (index / (numStops - 1)) * 100;
    return {
      value: stop.value,
      color: stop.color,
      positionPercent: 100 - percentageFromBottom,
      originalIndex: stop.originalIndex,
    };
  });
});
</script>

<style scoped>
.legend-container {
  position: fixed;
  right: 20px;
  bottom: 150px;
  padding: 10px;
  background-color: #f9f9f9;
  border: 1px solid #ccc;
  border-radius: 5px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  font-family: Arial, sans-serif;
  font-size: 12px;
}

.legend-title {
  font-weight: bold;
  margin-bottom: 8px;
  text-align: center;
}

.legend-body {
  display: flex;
  align-items: flex-start;
}

.color-bar-wrapper {
  margin-right: 8px;
  height: v-bind(height);
  display: flex;
  align-items: center;
}

.color-bar {
  border: 1px solid #eee;
}

.labels-container {
  position: relative;
  height: v-bind(height);
  flex-grow: 1;
}

.label-item {
  position: absolute;
  left: 0;
  width: 100%;
  transform: translateY(-50%);
  white-space: nowrap;
}

.label-text {
  display: inline-block;
  padding: 2px 0;
}
</style>
