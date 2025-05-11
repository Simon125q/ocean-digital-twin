<template>
  <div ref="mapContainer" class="map-container"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import mapboxgl, {Map, NavigationControl} from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';

const mapboxAccessToken: string | undefined = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;

if (!mapboxAccessToken) {
  console.error(
    'Mapbox access token is missing. Please set VITE_MAPBOX_ACCESS_TOKEN in your .env file.'
  );
}

const mapContainer: Ref<HTMLDivElement | null> = ref(null);
let mapInstance: Map | null = null;

onMounted(() => {
  if (!mapboxAccessToken) {
    console.warn('Mapbox access token not found. Map will not initialize.')
    return;
  }

  if (mapContainer.value) {
    mapboxgl.accessToken = mapboxAccessToken;
    mapInstance = new Map({
      container: mapContainer.value, // container ID or HTML element
      style: 'mapbox://styles/mapbox/streets-v12', // style URL
      center: [1.6, 41.16], // starting position [lng, lat]
      zoom: 10, // starting zoom
    });

    mapInstance.addControl(new mapboxgl.NavigationControl(), 'top-right');

    mapInstance.on('load', () => {
      console.log('Map loaded!');
      //TODO: load initial data
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
