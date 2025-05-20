<template>
  <div class="map-wrapper">
    <div ref="mapContainer" class="map-container"></div>
    <DataTypeSelector v-model="selectedDataType"/>
    <TimelineSlider
        :availableDates="availableDates"
        :onChange="handleDateChange"
        :key="availableDates.length"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, Ref, computed } from 'vue';
import mapboxgl, {Map as MapboxMap, NavigationControl, GeoJSONSource } from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import { fetchChlorophyllData } from '@/services/chlorophyllService';
import { fetchCurrentsData } from '@/services/currentsService';
import type { ChlorophyllFeatureCollection, ChlorophyllFeatureProperties } from '@/types/chlorophyll';
import type { CurrentsFeatureCollection, CurrentsFeatureProperties } from '@/types/currents';
import TimelineSlider from './TimelineSlider.vue'
import DataTypeSelector from './DataTypeSelector.vue';
import type { DataType } from './DataTypeSelector.vue';

const mapboxAccessToken: string | undefined = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;

if (!mapboxAccessToken) {
  console.error(
    'Mapbox access token is missing. Please set VITE_MAPBOX_ACCESS_TOKEN in your .env file.'
  );
}

const mapContainer: Ref<HTMLDivElement | null> = ref(null);
let mapInstance: MapboxMap | null = null;

const SOURCE_ID = 'chlorophyll-source';
const LAYER_ID = 'chlorophyll-layer';

const allChlorophyllData = ref<ChlorophyllFeatureCollection>({
  type: 'FeatureCollection',
  feature: []
})

const selectedDataType = ref<DataType>('chlorophyll');

const availableDates = ref<Date[]>([]);

const selectedDate = ref<Date | null>(null);

const filteredChlorophyllData = computed(() => {
  if (!selectedDate.value || allChlorophyllData.value.features.length === 0) {
    return {
      type: 'FeatureCollection',
      features: []
    } as ChlorophyllFeatureCollection;
  }
  const selectedDateStr = selectedDate.value.toISOString().split('T')[0];
  const filteredFeatures = allChlorophyllData.value.features.filter(feature => {
    const featureData = new Date(feature.properties.measurement_time);
    const featureDateStr = featureData.toISOString().split('T')[0];
    return featureDateStr === selectedDateStr;
  });
  return {
    type: 'FeatureCollection',
    features: filteredFeatures
  } as ChlorophyllfeatureCollection;
});

function extractAvailableDates(data: ChlorophyllFeatureCollection): Date[] {
  const dateMap = new Map<string, Date>();

  data.features.forEach(feature => {
    const dateStr = feature.properties.measurement_time.split('T')[0]; // Get just the date part
    if (!dateMap.has(dateStr)) {
      dateMap.set(dateStr, new Date(feature.properties.measurement_time));
    }
  });

  // Sort dates chronologically
  return Array.from(dateMap.values()).sort((a, b) => a.getTime() - b.getTime());
}

function handleDateChange(date: Date) {
  selectedDate.value = date;
  updateMapData();
}

function updateMapData() {
  if (!mapInstance || !mapInstance.getSource(SOURCE_ID)) return;

  (mapInstance.getSource(SOURCE_ID) as GeoJSONSource).setData(filteredChlorophyllData.value);
}

async function loadChlorophyllData(map: MapboxMap) {
  try {
    const chlorophyllGeoJson = await fetchChlorophyllData(true);
    allChlorophyllData.value = chlorophyllGeoJson;
    availableDates.value = extractAvailableDates(chlorophyllGeoJson);

    if (availableDates.value.length > 0) {
      selectedDate.value = availableDates.value[availableDates.value.length - 1];
    }

    if (map.getSource(SOURCE_ID)) {
      (map.getSource(SOURCE_ID) as GeoJSONSource).setData(chlorophyllGeoJson);
      console.log('Chlorophyll data source updated')
    } else {
      map.addSource(SOURCE_ID, {
        type: 'geojson',
        data: filteredChlorophyllData.value,
      });
      console.log('Chlorophyll data source added')

      map.addLayer({
        id: LAYER_ID,
        type: 'circle',
        source: SOURCE_ID,
        paint: {
          'circle-radius': [
            'interpolate',
            ['linear'],
            ['get', 'chlor_a'],
            0, 4,
            0.3, 8,
            1, 12
          ],
          'circle-color': [
            'interpolate',
            ['linear'],
            ['get', 'chlor_a'],
            0, '#ffffcc',
            0.3, '#41b6c4',
            1, '#0c2c84'
          ],
          'circle-opacity': 0.8,
          'circle-stroke-width': 1,
          'circle-stroke-color': '#ffffff'
        },
      });
      console.log('Chlorophyll data layer added');

      map.on('click', LAYER_ID, (e) => {
        if (e.features && e.features.length > 0) {
          const feature = e.features[0];
          const coordinates = (feature.geometry as GeoJSON.Point).coordinates.slice();
          const properties = feature.properties as ChlorophyllFeatureProperties;

          const chlorAValue = typeof properties.chlor_a === 'number' ?
          properties.chlor_a.toFixed(2) : 'N/A';

          const description =
          `
            <strong>Chlorophyll Data</strong><br>
            ID: ${properties.id}<br>
            Chlorophyll A: ${chlorAValue} µg/L<br>
            Time: ${new Date(properties.measurement_time).toLocaleString()}
          `;

          while (Math.abs(e.lngLat.lng - coordinates[0]) > 100) {
            coordinates[0] += e.lngLat.lng > coordinates[0] ? 360 : -360;
          }

          new mapboxgl.Popup()
            .setLngLat(coordinates as [number, number])
            .setHTML(description)
            .addTo(map);
        }
      });

      map.on('mouseenter', LAYER_ID, () => {
        map.getCanvas().style.cursor = 'pointer';
      });

      map.on('mouseleave', LAYER_ID, () => {
        map.getCanvas().style.cursor = '';
      });
    }
  } catch (error) {
    console.error("failed to load or display chlorophyll data:", error);
  }
}

onMounted(() => {
  if (!mapboxAccessToken) {
    console.warn('Mapbox access token not found. Map will not initialize.')
    return;
  }

  if (mapContainer.value) {
    mapboxgl.accessToken = mapboxAccessToken;
    mapInstance = new MapboxMap({
      container: mapContainer.value, // container ID or HTML element
      style: 'mapbox://styles/mapbox/streets-v12', // style URL
      center: [1.72, 41.00], // starting position [lng, lat]
      zoom: 10, // starting zoom
    });

    mapInstance.addControl(new mapboxgl.NavigationControl(), 'top-right');

    mapInstance.on('load', () => {
      console.log('Map loaded!');
      //TODO: load initial data
      loadChlorophyllData(mapInstance);
    });

    mapInstance.on('error', (e) => {
      console.error('mapbox error:', e);
    });
  } else {
    console.error('Map container elem not found.')
  }
});

onUnmounted(() => {
  if (mapInstance) {
    mapInstance.remove();
    mapInstance = null;
    console.log('Map instance removed');
  }
});
</script>

<style scoped>
.map-container {
  width: 100%;
  height: 100vh;
}
</style>
