<template>
  <div>
    <label class="block text-sm font-medium text-gray-700">Targets</label>
    <div class="mt-2 space-y-4">
      <div v-for="(target, index) in modelValue" :key="index" class="flex items-center space-x-2">
        <input
            v-model="target.url"
            type="text"
            placeholder="Target URL"
            class="block p-2 w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        />
        <input
            v-model="target.name"
            type="text"
            placeholder="Target Name"
            class="block p-2 w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        />
        <input
            v-model.number="target.weight"
            type="number"
            min="0"
            max="100"
            placeholder="Weight"
            class="block p-2 w-20 rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        />
        <button
            type="button"
            @click="removeTarget(index)"
            class="text-red-600 hover:text-red-900"
        >
          Remove
        </button>
      </div>
      <button
          type="button"
          @click="addTarget"
          class="mt-2 text-sm text-indigo-600 hover:text-indigo-900"
      >
        Add Target
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue';

type Target = {
  id?: string;
  name?: string;
  url: string;
  weight: number;
  is_active?: boolean
};

interface Props {
  modelValue: Target[];
}

const props = defineProps<Props>();
const emit = defineEmits(['update:modelValue', 'add', 'remove']);

const addTarget = () => {
  const newTargets = [...props.modelValue, { url: '', weight: 100, name: '' }];
  emit('update:modelValue', newTargets);
};

const removeTarget = (index: number) => {
  const newTargets = [...props.modelValue];
  newTargets.splice(index, 1);
  emit('update:modelValue', newTargets);
};
</script>
