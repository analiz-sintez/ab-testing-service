<template>
  <div class="space-y-3">
    <h3 class="text-sm font-medium text-gray-700">Settings</h3>
    
    <div class="flex flex-wrap gap-4">
      <!-- Save Cookies checkbox -->
      <div class="relative flex items-center group">
        <input
            id="saving_cookies_flg"
            v-model="settings.saving_cookies_flg"
            type="checkbox"
            class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
            @change="updateSettings"
        />
        <label for="saving_cookies_flg" class="ml-2 text-sm font-medium text-gray-700 cursor-pointer">
          Save Cookies
          <span class="ml-1 text-gray-400 hover:text-gray-600">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </span>
        </label>
        <!-- Tooltip -->
        <div class="absolute left-0 bottom-full mb-2 hidden group-hover:block z-10">
          <div class="bg-gray-800 text-white text-xs rounded py-1 px-2 w-48">
            Enable cookie persistence across requests
            <svg class="absolute text-gray-800 h-2 w-full left-0 top-full" x="0px" y="0px" viewBox="0 0 255 255" xml:space="preserve">
              <polygon class="fill-current" points="0,0 127.5,127.5 255,0" />
            </svg>
          </div>
        </div>
      </div>

      <!-- Query Forwarding checkbox -->
      <div class="relative flex items-center group">
        <input
            id="query_forwarding_flg"
            v-model="settings.query_forwarding_flg"
            type="checkbox"
            class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
            @change="updateSettings"
        />
        <label for="query_forwarding_flg" class="ml-2 text-sm font-medium text-gray-700 cursor-pointer">
          Forward Query Parameters
          <span class="ml-1 text-gray-400 hover:text-gray-600">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </span>
        </label>
        <!-- Tooltip -->
        <div class="absolute left-0 bottom-full mb-2 hidden group-hover:block z-10">
          <div class="bg-gray-800 text-white text-xs rounded py-1 px-2 w-48">
            Forward query parameters to target URLs
            <svg class="absolute text-gray-800 h-2 w-full left-0 top-full" x="0px" y="0px" viewBox="0 0 255 255" xml:space="preserve">
              <polygon class="fill-current" points="0,0 127.5,127.5 255,0" />
            </svg>
          </div>
        </div>
      </div>

      <!-- Cookies Forwarding checkbox -->
      <div class="relative flex items-center group">
        <input
            id="cookies_forwarding_flg"
            v-model="settings.cookies_forwarding_flg"
            type="checkbox"
            class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
            @change="updateSettings"
        />
        <label for="cookies_forwarding_flg" class="ml-2 text-sm font-medium text-gray-700 cursor-pointer">
          Forward Cookies
          <span class="ml-1 text-gray-400 hover:text-gray-600">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </span>
        </label>
        <!-- Tooltip -->
        <div class="absolute left-0 bottom-full mb-2 hidden group-hover:block z-10">
          <div class="bg-gray-800 text-white text-xs rounded py-1 px-2 w-48">
            Forward cookies to target URLs
            <svg class="absolute text-gray-800 h-2 w-full left-0 top-full" x="0px" y="0px" viewBox="0 0 255 255" xml:space="preserve">
              <polygon class="fill-current" points="0,0 127.5,127.5 255,0" />
            </svg>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';

interface Settings {
  saving_cookies_flg: boolean;
  query_forwarding_flg: boolean;
  cookies_forwarding_flg: boolean;
}

interface Props {
  modelValue: Settings;
}

const props = defineProps<Props>();
const emit = defineEmits(['update:modelValue']);

// Create a local copy of the settings to work with
const settings = ref<Settings>({ ...props.modelValue });

// Watch for external changes to the modelValue
watch(() => props.modelValue, (newVal) => {
  settings.value = { ...newVal };
}, { deep: true });

// Update the parent component when local settings change
const updateSettings = () => {
  emit('update:modelValue', { ...settings.value });
};
</script>
