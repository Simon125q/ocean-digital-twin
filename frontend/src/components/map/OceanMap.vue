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
import { ref, onMounted, onUnmounted, Ref, computed, watch } from 'vue';
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

const DATA_SOURCES = {
  chlorophyll: 'chlorophyll-source',
  v_current: 'v-current-source',
  u_current: 'u-current-source',
  combined_current: 'combined-current-source'
};

const DATA_LAYERS = {
  chlorophyll: 'chlorophyll-layer',
  v_current: 'v-current-layer',
  u_current: 'u-current-layer',
  combined_current: 'combined-current-layer'
};

const allChlorophyllData = ref<ChlorophyllFeatureCollection>({
  type: 'FeatureCollection',
  feature: []
})

const allCurrentsData = ref<CurrentsFeatureCollection>({
  type: 'FeatureCollection',
  features: []
});

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

const filteredCurrentsData = computed(() => {
  if (!selectedDate.value || allCurrentsData.value.features.length === 0) {
    return {
      type: 'FeatureCollection',
      features: []
    } as CurrentsFeatureCollection;
  }

  const selectedDateStr = selectedDate.value.toISOString().split('T')[0];
  const filteredFeatures = allCurrentsData.value.features.filter(feature => {
    const featureDate = new Date(feature.properties.measurement_time);
    const featureDateStr = featureDate.toISOString().split('T')[0];
    return featureDateStr === selectedDateStr;
  });

  return {
    type: 'FeatureCollection',
    features: filteredFeatures
  } as CurrentsFeatureCollection;
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
  if (!mapInstance) return;

  // Update the visible layer based on selected data type
  Object.values(DATA_LAYERS).forEach(layerId => {
    if (mapInstance?.getLayer(layerId)) {
      mapInstance.setLayoutProperty(
        layerId,
        'visibility',
        layerId === DATA_LAYERS[selectedDataType.value] ? 'visible' : 'none'
      );
    }
  });

  // Update the data for the selected source
  const sourceId = DATA_SOURCES[selectedDataType.value];
  if (mapInstance.getSource(sourceId)) {
    if (selectedDataType.value === 'chlorophyll') {
      (mapInstance.getSource(sourceId) as GeoJSONSource).setData(filteredChlorophyllData.value);
    } else {
      (mapInstance.getSource(sourceId) as GeoJSONSource).setData(filteredCurrentsData.value);
    }
  }
}

const squareSvgString =
  '<svg width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><rect width="24" height="24" fill="black"/></svg>';
const northArrowSvgString =
  '<svg width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M12 2L4 14h16L12 2z" fill="black"/></svg>';
const southArrowSvgString =
  '<svg width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M12 22L4 10h16L12 22z" fill="black"/></svg>';
const eastArrowSvgString =
  '<svg width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M22 12L10 4v16L22 12z" fill="black"/></svg>';
const westArrowSvgString =
  '<svg width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M2 12L14 4v16L2 12z" fill="black"/></svg>';
const combinedArrowSvgString =
  '<svg width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M12 2L16 8H8L12 2z M22 12L16 16V8L22 12z M12 22L8 16H16L12 22z M2 12L8 8V16L2 12z" fill="black"/></svg>';

const squareImg = new Image(24, 24);
squareImg.onerror = (e) => console.error("Error loading SVG for square icon:", e);
squareImg.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(squareSvgString);

const northArrowImg = new Image(24, 24);
northArrowImg.onerror = (e) => console.error("Error loading SVG for north arrow icon:", e);
northArrowImg.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(northArrowSvgString);

const southArrowImg = new Image(24, 24);
southArrowImg.onerror = (e) => console.error("Error loading SVG for south arrow icon:", e);
southArrowImg.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(southArrowSvgString);

const eastArrowImg = new Image(24, 24);
eastArrowImg.onerror = (e) => console.error("Error loading SVG for east arrow icon:", e);
eastArrowImg.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(eastArrowSvgString);

const westArrowImg = new Image(24, 24);
westArrowImg.onerror = (e) => console.error("Error loading SVG for west arrow icon:", e);
westArrowImg.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(westArrowSvgString);

const combinedArrowImg = new Image(24, 24);
combinedArrowImg.onerror = (e) => console.error("Error loading SVG for combined arrow icon:", e);
combinedArrowImg.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(combinedArrowSvgString);

async function loadChlorophyllData(map: MapboxMap) {
  try {
    const chlorophyllGeoJson = await fetchChlorophyllData(true);
    allChlorophyllData.value = chlorophyllGeoJson;
    availableDates.value = extractAvailableDates(chlorophyllGeoJson);

    if (availableDates.value.length > 0) {
      selectedDate.value = availableDates.value[availableDates.value.length - 1];
    }

    if (map.getSource(DATA_SOURCES.chlorophyll)) {
      (map.getSource(DATA_SOURCES.chlorophyll) as GeoJSONSource).setData(chlorophyllGeoJson);
      console.log('Chlorophyll data source updated')
    } else {
      map.addSource(DATA_SOURCES.chlorophyll, {
        type: 'geojson',
        data: filteredChlorophyllData.value,
      });
      console.log('Chlorophyll data source added')

      map.addLayer({
      id: DATA_LAYERS.chlorophyll,
      type: 'symbol',
      source: DATA_SOURCES.chlorophyll,
      layout: {
        'icon-image': 'chlorophyll-square',
        'icon-allow-overlap': false,
        'icon-ignore-placement': true,
        'icon-size': [
          'interpolate',
          ['linear'],
          ['zoom'],
          5,  [
            '*',
            [
              'interpolate',
              ['linear'],
              ['get', 'chlor_a'],
              0, 1,
              1, 1,
            ],
            0.5
          ],
          10, [
            '*',
            [
              'interpolate',
              ['linear'],
              ['get', 'chlor_a'],
              0, 1,
              1, 1,
            ],
            3.5
          ],
          15, [
            '*',
            [
              'interpolate',
              ['linear'],
              ['get', 'chlor_a'],
              0, 1,
              1, 1,
            ],
            15.0
          ]
        ],
        'visibility': selectedDataType.value === 'chlorophyll' ? 'visible' : 'none',
      },
      paint: {
        'icon-color': [
          'interpolate',
          ['linear'],
          ['get', 'chlor_a'],
          0, '#ffffcc',
          0.3, '#41b6c4',
          1, '#0c2c84'
        ],
        'icon-opacity': 1.0,
      },
    });

      console.log('Chlorophyll data layer added');

      map.on('click', DATA_LAYERS.chlorophyll, (e) => {
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

      map.on('mouseenter', DATA_LAYERS.chlorophyll, () => {
        map.getCanvas().style.cursor = 'pointer';
      });

      map.on('mouseleave', DATA_LAYERS.chlorophyll, () => {
        map.getCanvas().style.cursor = '';
      });
    }
  } catch (error) {
    console.error("failed to load or display chlorophyll data:", error);
  }
}

async function loadCurrentsData(map: MapboxMap) {
  try {
    const currentsGeoJson = await fetchCurrentsData(true);
    allCurrentsData.value = currentsGeoJson;

    if (availableDates.value.length === 0) {
      availableDates.value = extractAvailableDates(currentsGeoJson);
      if (availableDates.value.length > 0) {
        selectedDate.value = availableDates.value[availableDates.value.length - 1];
      }
    }

    setupCurrentsLayer(map, 'v_current', 'v-current', (value) => value > 0 ? 'north-arrow' : 'south-arrow');
    setupCurrentsLayer(map, 'u_current', 'u-current', (value) => value > 0 ? 'east-arrow' : 'west-arrow');
    setupCombinedCurrentsLayer(map);

  } catch (error) {
    console.error("Failed to load or display currents data:", error);
  }
}

function setupCurrentsLayer(
  map: MapboxMap,
  dataType: 'v_current' | 'u_current',
  layerPrefix: string,
  getIconImage: (value: number) => string
) {
  const sourceId = DATA_SOURCES[dataType];
  const layerId = DATA_LAYERS[dataType];

  if (map.getSource(sourceId)) {
    (map.getSource(sourceId) as GeoJSONSource).setData(filteredCurrentsData.value);
    console.log(`${layerPrefix} data source updated`);
  } else {
    map.addSource(sourceId, {
      type: 'geojson',
      data: filteredCurrentsData.value,
    });
    console.log(`${layerPrefix} data source added`);

    map.addLayer({
      id: layerId,
      type: 'symbol',
      source: sourceId,
      layout: {
        'icon-image': [
          'case',
          ['>', ['get', dataType], 0],
          getIconImage(1),
          getIconImage(-1)
        ],
        'icon-allow-overlap': false,
        'icon-ignore-placement': true,
        'icon-size': [
          'interpolate',
          ['linear'],
          ['zoom'],
          5, [
            '*',
            [
              'interpolate',
              ['linear'],
              ['abs', ['get', dataType]],
              0, 0.5,
              0.5, 1.5,
              1, 2.5
            ],
            2
          ],
          10, [
            '*',
            [
              'interpolate',
              ['linear'],
              ['abs', ['get', dataType]],
              0, 0.5,
              0.5, 1.5,
              1, 2.5
            ],
            6
          ],
          15, [
            '*',
            [
              'interpolate',
              ['linear'],
              ['abs', ['get', dataType]],
              0, 0.5,
              0.5, 1.5,
              1, 2.5
            ],
            9.0
          ]
        ],
        'visibility': selectedDataType.value === dataType ? 'visible' : 'none',
      },
      paint: {
        'icon-color': [
          'interpolate',
          ['linear'],
          ['abs', ['get', dataType]],
          0, '#ffffcc',
          0.3, '#41b6c4',
          1, '#0c2c84'
        ],
        'icon-opacity': 1.0,
      },
    });

    console.log(`${layerPrefix} data layer added`);

    map.on('click', layerId, (e) => {
      if (e.features && e.features.length > 0) {
        const feature = e.features[0];
        const coordinates = (feature.geometry as GeoJSON.Point).coordinates.slice();
        const properties = feature.properties as CurrentsFeatureProperties;

        const currentValue = typeof properties[dataType] === 'number' ?
          properties[dataType].toFixed(2) : 'N/A';

        const description =
          `
            <strong>${dataType === 'v_current' ? 'North-South' : 'West-East'} Current Data</strong><br>
            ID: ${properties.id}<br>
            Current: ${currentValue} m/s<br>
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

    map.on('mouseenter', layerId, () => {
      map.getCanvas().style.cursor = 'pointer';
    });

    map.on('mouseleave', layerId, () => {
      map.getCanvas().style.cursor = '';
    });
  }
}

function setupCombinedCurrentsLayer(map: MapboxMap) {
  const sourceId = DATA_SOURCES.combined_current;
  const layerId = DATA_LAYERS.combined_current;

  if (map.getSource(sourceId)) {
    (map.getSource(sourceId) as GeoJSONSource).setData(filteredCurrentsData.value);
    console.log('Combined currents data source updated');
  } else {
    map.addSource(sourceId, {
      type: 'geojson',
      data: filteredCurrentsData.value,
    });
    console.log('Combined currents data source added');

    map.addLayer({
      id: layerId,
      type: 'symbol',
      source: sourceId,
      layout: {
        'icon-image': 'combined-arrow',
        'icon-allow-overlap': false,
        'icon-ignore-placement': true,
        'icon-rotate': [
          'case',
          ['==', ['get', 'u_current'], 0],
          ['case',
            ['>', ['get', 'v_current'], 0],
            0,  // North
            180 // South
          ],
          [
            'case',
            ['==', ['get', 'v_current'], 0],
            ['case',
              ['>', ['get', 'u_current'], 0],
              90,  // East
              270  // West
            ],
            [
              '+',
              ['*',
                ['atan', ['/', ['get', 'v_current'], ['get', 'u_current']]],
                57.2958  // Convert radians to degrees
              ],
              ['case', ['<', ['get', 'u_current'], 0], 180, 0]
            ]
          ]
        ],
        'icon-size': [
          'interpolate',
          ['linear'],
          ['zoom'],
          5, [
            '*',
            [
              'interpolate',
              ['linear'],
              [
                'sqrt',
                ['+',
                  ['*', ['get', 'u_current'], ['get', 'u_current']],
                  ['*', ['get', 'v_current'], ['get', 'v_current']]
                ]
              ],
              0, 0.5,
              0.5, 1.5,
              1, 2.5
            ],
            3
          ],
          10, [
            '*',
            [
              'interpolate',
              ['linear'],
              [
                'sqrt',
                ['+',
                  ['*', ['get', 'u_current'], ['get', 'u_current']],
                  ['*', ['get', 'v_current'], ['get', 'v_current']]
                ]
              ],
              0, 0.5,
              0.5, 1.5,
              1, 2.5
            ],
            6
          ],
          15, [
            '*',
            [
              'interpolate',
              ['linear'],
              [
                'sqrt',
                ['+',
                  ['*', ['get', 'u_current'], ['get', 'u_current']],
                  ['*', ['get', 'v_current'], ['get', 'v_current']]
                ]
              ],
              0, 0.5,
              0.5, 1.5,
              1, 2.5
            ],
            9
          ]
        ],
        'visibility': selectedDataType.value === 'combined_current' ? 'visible' : 'none',
      },
      paint: {
        'icon-color': [
          'interpolate',
          ['linear'],
          [
            'sqrt',
            ['+',
              ['*', ['get', 'u_current'], ['get', 'u_current']],
              ['*', ['get', 'v_current'], ['get', 'v_current']]
            ]
          ],
          0, '#ffffcc',
          0.3, '#41b6c4',
          1, '#0c2c84'
        ],
        'icon-opacity': 1.0,
      },
    });

    console.log('Combined currents data layer added');

    map.on('click', layerId, (e) => {
      if (e.features && e.features.length > 0) {
        const feature = e.features[0];
        const coordinates = (feature.geometry as GeoJSON.Point).coordinates.slice();
        const properties = feature.properties as CurrentsFeatureProperties;

        const uCurrentValue = typeof properties.u_current === 'number' ?
          properties.u_current.toFixed(2) : 'N/A';
        const vCurrentValue = typeof properties.v_current === 'number' ?
          properties.v_current.toFixed(2) : 'N/A';

        const magnitude = Math.sqrt(
          Math.pow(properties.u_current, 2) + Math.pow(properties.v_current, 2)
        ).toFixed(2);

        const description =
          `
            <strong>Combined Current Data</strong><br>
            ID: ${properties.id}<br>
            West-East: ${uCurrentValue} m/s<br>
            North-South: ${vCurrentValue} m/s<br>
            Magnitude: ${magnitude} m/s<br>
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

    map.on('mouseenter', layerId, () => {
      map.getCanvas().style.cursor = 'pointer';
    });

    map.on('mouseleave', layerId, () => {
      map.getCanvas().style.cursor = '';
    });
  }
}

watch(selectedDataType, () => {
  updateMapData();
});

onMounted(() => {
  if (!mapboxAccessToken) {
    console.warn('Mapbox access token not found. Map will not initialize.')
    return;
  }

  if (mapContainer.value) {
    mapboxgl.accessToken = mapboxAccessToken;
    mapInstance = new MapboxMap({
      container: mapContainer.value, // container ID or HTML element
      style: 'mapbox://styles/mapbox/outdoors-v12', // style URL
      center: [1.72, 41.00], // starting position [lng, lat]
      zoom: 10, // starting zoom
    });

    mapInstance.addControl(new mapboxgl.NavigationControl(), 'top-right');

    mapInstance.addImage('chlorophyll-square', squareImg, { sdf: true });
    mapInstance.addImage('north-arrow', northArrowImg, { sdf: true });
    mapInstance.addImage('south-arrow', southArrowImg, { sdf: true });
    mapInstance.addImage('east-arrow', eastArrowImg, { sdf: true });
    mapInstance.addImage('west-arrow', westArrowImg, { sdf: true });
    mapInstance.addImage('combined-arrow', combinedArrowImg, { sdf: true });

    mapInstance.on('load', () => {
      console.log('Map loaded!');
      loadChlorophyllData(mapInstance);
      loadCurrentsData(mapInstance);
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
