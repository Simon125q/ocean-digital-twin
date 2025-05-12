<template>
  <div ref="mapContainer" class="map-container"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import mapboxgl, {Map, NavigationControl, GeoJSONSource } from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import { fetchChlorophyllData } from '@/services/chlorophyllService';
import type { ChlorophyllFeatureCollection } from '@/types/chlorophyll';

const mapboxAccessToken: string | undefined = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;

if (!mapboxAccessToken) {
  console.error(
    'Mapbox access token is missing. Please set VITE_MAPBOX_ACCESS_TOKEN in your .env file.'
  );
}

const mapContainer: Ref<HTMLDivElement | null> = ref(null);
let mapInstance: Map | null = null;

const SOURCE_ID = 'chlorophyll-source';
const LAYER_ID = 'chlorophyll-layer';

async function loadChlorophyllData(map: Map) {
  try {
    const chlorophyllGeoJson = await fetchChlorophyllData();
    if (map.getSource(SOURCE_ID)) {
      (map.getSource(SOURCE_ID) as GeoJSONSource).setData(chlorophyllGeoJson);
      console.log('Chlorophyll data source updated')
    } else {
      map.addSource(SOURCE_ID, {
        type: 'geojson',
        data: chlorophyllGeoJson,
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
            1, 15
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
          //TODO:
          const coordinates = (feature.geometry as GeoJSON.Point).coordinates.slice();
          const properties = feature.properties as ChlorophyllFeatureProperties;

          //TODO:
          const chlorAValue = typeof properties.chlor_a === 'number' ?
          properties.chlor_a.toFixed(2) : 'N/A';

          //TODO:
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
