<template>
  <div>
    <label class="block text-sm font-medium text-gray-700">Listen URLs</label>
    <div class="flex items-center">
      <span class="text-xs text-gray-500 mr-1">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </span>
      <span class="text-xs text-gray-500">Multiple URLs can be used for the same proxy</span>
    </div>
    <div class="mt-2 space-y-2">
      <div v-for="(url, index) in modelValue" :key="index" class="flex items-center space-x-2">
        <div class="relative flex-grow">
          <input
              :value="url"
              @input="updateUrl(index, ($event.target as HTMLInputElement).value)"
              @blur="validateUrl(index)"
              type="text"
              placeholder="Enter listen URL (e.g., localhost:8080)"
              class="block w-full p-2 rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
              :class="{ 'border-red-300': errors[index] }"
          />
          <div v-if="index === 0" class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
            <span class="text-xs font-medium text-gray-500 bg-gray-100 px-2 py-1 rounded-full">
              Primary
              <span class="ml-1 cursor-help" title="This is the primary URL that will be used for this proxy">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </span>
            </span>
          </div>
          <p v-if="errors[index]" class="mt-1 text-sm text-red-600">{{ errors[index] }}</p>
        </div>
        <button
            v-if="modelValue.length > 1"
            type="button"
            @click="removeUrl(index)"
            class="text-red-600 hover:text-red-900"
        >
          Remove
        </button>
      </div>
      <button
          type="button"
          @click="addUrl"
          class="flex items-center justify-center w-full py-2 border border-dashed border-gray-300 rounded-md text-sm text-indigo-600 hover:text-indigo-900 hover:bg-indigo-50 transition-colors"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Add Another Listen URL
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue';

interface Props {
  modelValue: string[];
  errors: Record<number, string>;
}

const props = defineProps<Props>();
const emit = defineEmits(['update:modelValue', 'validate', 'add', 'remove']);

const updateUrl = (index: number, value: string) => {
  const newUrls = [...props.modelValue];
  newUrls[index] = value;
  emit('update:modelValue', newUrls);
};

const validateUrl = (index: number) => {
  emit('validate', index);
};

const addUrl = () => {
  emit('add');
};

const removeUrl = (index: number) => {
  emit('remove', index);
};
</script>
