<template>
  <div class="timeline-container">
    <div class="timeline-controls">
      <button
        class="timeline-button"
        @click="togglePlay"
        :title="isPlaying ? 'Pause' : 'Play'"
      >
        {{ isPlaying ? '⏸' : '▶️' }}
      </button>

      <div class="timeline-date">
        {{ formatDate(currentDate) }}
      </div>
    </div>

    <div class="timeline-slider-container">
      <input
        type="range"
        class="timeline-slider"
        :min="0"
        :max="availableDates.length - 1"
        :value="currentDateIndex"
        @input="onSliderChange"
      />
      <div class="timeline-labels">
        <span class="timeline-label-start">{{ formatDate(availableDates[0]) }}</span>
        <span class="timeline-label-end">{{ formatDate(availableDates[availableDates.length - 1]) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onUnmounted, onMounted, watch } from 'vue';

  interface TimelineProps {
    availableDates: Date[];
    onChange: (date: Date) => void;
  }

  const props = defineProps<TimelineProps>();

  const currentDateIndex = ref(props.availableDates.length - 1);
  const isPlaying = ref(false);
  let animationInterval: number | null = null;

  const currentDate = computed(() => {
    if (props.availableDates.length === 0) return new Date();
    return props.availableDates[currentDateIndex.value];
  });

  function formatDate(date: Date): string {
    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'short',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  }

  function onSliderChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const newIndex = parseInt(target.value, 10);
    currentDateIndex.value = newIndex;
    props.onChange(props.availableDates[newIndex]);
  }

  function togglePlay() {
    isPlaying.value = !isPlaying.value;

    if (isPlaying.value) {
      startAnimation();
    } else if (animationInterval !== null) {
      clearInterval(animationInterval);
      animationInterval = null;
    }
  }

  function startAnimation() {
    if (animationInterval !== null) {
      clearInterval(animationInterval);
    }

    animationInterval = window.setInterval(() => {
      if (currentDateIndex.value >= props.availableDates.length - 1) {
        currentDateIndex.value = 0;
      } else {
        currentDateIndex.value++;
      }
      props.onChange(props.availableDates[currentDateIndex.value]);
    }, 1000)
  }

onMounted(() => {
  if (props.availableDates.length > 0) {
    currentDateIndex.value = props.availableDates.length - 1;
    props.onChange(props.availableDates[currentDateIndex.value]);
  }
});

  watch(() => props.availableDates, (newDates) => {
    if (newDates.length > 0 && currentDateIndex.value >= newDates.length) {
      currentDateIndex.value = newDates.length - 1;
      props.onChange(newDates[currentDateIndex.value]);
    }
  }, { deep: true });

  onUnmounted(() => {
    if (animationInterval !== null) {
      clearInterval(animationInterval);
      animationInterval = null;
    }
  });
</script>

<style scoped>
.timeline-container {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  width: 80%;
  max-width: 800px;
  background-color: rgba(255, 255, 255, 0.9);
  border-radius: 8px;
  padding: 10px 15px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
  z-index: 1;
}

.timeline-controls {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.timeline-button {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  padding: 0 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.timeline-date {
  font-weight: bold;
  margin-left: 10px;
  flex-grow: 1;
  text-align: center;
}

.timeline-slider-container {
  position: relative;
}

.timeline-slider {
  width: 100%;
  height: 8px;
  -webkit-appearance: none;
  appearance: none;
  background: #d3d3d3;
  outline: none;
  border-radius: 4px;
}

.timeline-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #2196F3;
  cursor: pointer;
}

.timeline-slider::-moz-range-thumb {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #2196F3;
  cursor: pointer;
}

.timeline-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 5px;
  font-size: 12px;
  color: #555;
}
</style>
