<template>
  <div>
    <label class="block text-sm font-medium text-gray-700">Route Condition</label>
    <div class="mt-2 space-y-4">
      <div>
        <select
            v-model="condition.type"
            class="mt-1 p-2 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            @change="updateCondition"
        >
          <option value="">No condition</option>
          <option value="header">Header</option>
          <option value="query">Query Parameter</option>
          <option value="cookie">Cookie</option>
          <option value="user_agent">User Agent</option>
          <option value="language">Language</option>
          <option value="expr">Expression</option>
        </select>
        <p v-if="condition.type === 'language'" class="mt-1 text-sm text-gray-500">
          {{ getConditionHelp('language', '') }}
        </p>
        <p v-if="condition.type === 'user_agent'" class="mt-1 text-sm text-gray-500">
          {{ getConditionHelp('user_agent', condition.param_name || '') }}
        </p>
      </div>

      <!-- Parameter Name -->
      <div v-if="condition.type && condition.type !== 'language' && condition.type !== 'expr'">
        <label class="block text-sm font-medium text-gray-700">Parameter Name</label>
        <select
            v-if="condition.type === 'user_agent'"
            v-model="condition.param_name"
            class="mt-1 p-2 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        >
          <option value="platform">Platform (Mobile/Desktop)</option>
          <option value="browser">Browser</option>
        </select>
        <input
            v-else
            v-model="condition.param_name"
            type="text"
            :placeholder="getParamPlaceholder(condition.type)"
            class="mt-1 p-2 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            @input="updateCondition"
        />
        <p v-if="condition.param_name" class="mt-1 text-sm text-gray-500">
          {{ getConditionHelp(condition.type, condition.param_name) }}
        </p>
      </div>

      <!-- Single Expression -->
      <div v-if="condition.type === 'expr'">
        <label class="block text-sm font-medium text-gray-700">Expression Type</label>
        <div class="mt-2">
          <div class="flex items-center space-x-4">
            <div class="flex items-center">
              <input
                  id="single-expr"
                  type="radio"
                  :value="true"
                  v-model="useSingleExpression"
                  class="h-4 w-4 border-gray-300 text-indigo-600 focus:ring-indigo-500"
                  @change="toggleExpressionMode"
              />
              <label for="single-expr" class="ml-2 block text-sm text-gray-700">Single Expression</label>
            </div>
            <div class="flex items-center">
              <input
                  id="multiple-expr"
                  type="radio"
                  :value="false"
                  v-model="useSingleExpression"
                  class="h-4 w-4 border-gray-300 text-indigo-600 focus:ring-indigo-500"
                  @change="toggleExpressionMode"
              />
              <label for="multiple-expr" class="ml-2 block text-sm text-gray-700">Multiple Expressions</label>
            </div>
          </div>
        </div>
      </div>

      <!-- Single Expression Input -->
      <div v-if="condition.type === 'expr' && useSingleExpression">
        <label class="block text-sm font-medium text-gray-700">Single Expression</label>
        <textarea
            v-model="condition.expr"
            rows="3"
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            placeholder="headers['user-agent'] contains 'iPhone' ? 'target-1' : 'target-2'"
            @input="updateCondition"
        />
        <p class="mt-1 text-sm text-gray-500">
          Expression should return a target ID based on request properties
        </p>
      </div>
      
      <!-- Multiple Expressions -->
      <div v-if="condition.type === 'expr' && !useSingleExpression">
        <label class="block text-sm font-medium text-gray-700">Expressions</label>
        <p class="text-sm text-gray-500">{{ getConditionHelp('expr', '') }}</p>
        <div v-for="(expr, index) in condition.expressions" :key="index" class="mt-2">
          <div class="flex items-center space-x-2">
            <input
                v-model="expr.expr"
                type="text"
                placeholder="Expression (e.g., req.Header.Get('User-Agent') contains 'Mobile')"
                class="block p-2 w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                @input="updateCondition"
            />
            <select
                v-model="expr.target"
                class="block p-2 w-40 rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                @change="updateCondition"
            >
              <option
                  v-for="target in targets"
                  :key="target.id"
                  :value="target.id"
              >
                {{ target.name || target.url }}
              </option>
            </select>
            <button
                type="button"
                @click="removeExpression(index)"
                class="text-red-600 hover:text-red-900"
            >
              Remove
            </button>
          </div>
        </div>
        <button
            type="button"
            @click="addExpression"
            class="mt-2 text-sm text-indigo-600 hover:text-indigo-900"
        >
          Add Expression
        </button>
        <p class="mt-1 text-sm text-gray-500">
          Each expression should evaluate to a boolean. The first expression that evaluates to true will be used.
        </p>
      </div>

      <!-- Values -->
      <div v-if="condition.type && condition.type !== 'expr'">
        <label class="block text-sm font-medium text-gray-700">Values</label>
        <div class="mt-2 space-y-2">
          <div
              v-for="(target, index) in targets"
              :key="target.id || index"
              class="flex items-center gap-2"
          >
            <input
                v-model="condition.values[target.id || index]"
                type="text"
                class="block p-2 w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                :placeholder="getValuePlaceholder(condition.type, condition.param_name)"
                @input="updateCondition"
            />
            <span class="text-sm text-gray-500">→</span>
            <span class="text-sm text-gray-700">{{ target.name || target.url }}</span>
          </div>
        </div>
      </div>

      <!-- Default Target -->
      <div v-if="condition.type">
        <label class="block text-sm font-medium text-gray-700">Default Target</label>
        <select
            v-model="condition.default"
            class="mt-1 p-2 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            @change="updateCondition"
        >
          <option
              v-for="target in targets"
              :key="target.id"
              :value="target.id"
          >
            {{ target.name || target.url }}
          </option>
        </select>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, defineProps, defineEmits } from 'vue';

type Target = {
  id?: string;
  name?: string;
  url: string;
  weight: number;
  is_active?: boolean
};

type Expression = {
  expr: string;
  target: string;
};

type Condition = {
  type: string;
  param_name?: string;
  expressions?: Expression[];
  default?: string;
  values?: Record<string, string>;
  expr?: string; // For single expression mode
};

interface Props {
  modelValue: Condition;
  targets: Target[];
}

const props = defineProps<Props>();
const emit = defineEmits(['update:modelValue']);

// Create a local copy of the condition to work with
const condition = ref<Condition>({ ...props.modelValue });

// Track whether to use single expression or multiple expressions mode
const useSingleExpression = ref<boolean>(!!condition.value.expr);

// Update the parent component when local condition changes
const updateCondition = () => {
  emit('update:modelValue', { ...condition.value });
};

// Toggle between single and multiple expression modes
function toggleExpressionMode() {
  if (useSingleExpression.value) {
    // Switching to single expression mode
    condition.value.expressions = [];
    if (!condition.value.expr) {
      condition.value.expr = '';
    }
  } else {
    // Switching to multiple expressions mode
    condition.value.expr = '';
    if (!condition.value.expressions || condition.value.expressions.length === 0) {
      condition.value.expressions = [{ expr: '', target: props.targets[0]?.id || '' }];
    }
  }
  updateCondition();
}

// Watch for external changes to the modelValue
watch(() => props.modelValue, (newVal) => {
  condition.value = { ...newVal };
}, { deep: true });

// Watch for condition type changes to initialize appropriate properties
watch(() => condition.value.type, (newType) => {
  if (newType === 'expr') {
    // For expression type, initialize based on the selected mode
    if (useSingleExpression.value) {
      if (!condition.value.expr) {
        condition.value.expr = '';
      }
    } else {
      if (!condition.value.expressions || condition.value.expressions.length === 0) {
        condition.value.expressions = [{ expr: '', target: props.targets[0]?.id || '' }];
      }
    }
    updateCondition();
  } else if (newType && !condition.value.values) {
    // For other condition types, initialize values object
    condition.value.values = {};
    updateCondition();
  }
});

// Watch for changes in useSingleExpression to update the condition accordingly
watch(() => useSingleExpression.value, toggleExpressionMode);

// Initialize values object if it doesn't exist and condition type requires it
if (condition.value.type && condition.value.type !== 'expr' && !condition.value.values) {
  condition.value.values = {};
}

function getParamPlaceholder(type: string): string {
  switch (type) {
    case 'header':
      return 'Header name (e.g., User-Agent)';
    case 'query':
      return 'Query parameter name (e.g., version)';
    case 'cookie':
      return 'Cookie name (e.g., session_id)';
    case 'user_agent':
      return 'User agent pattern (e.g., Mobile)';
    default:
      return '';
  }
}

function getValuePlaceholder(type, param) {
  switch (type) {
    case 'header':
      return `Header value for ${param}`
    case 'query':
      return `Value for ${param} parameter`
    case 'cookie':
      return `Value for ${param} cookie`
    default:
      return ''
  }
}

function getConditionHelp(type, param) {
  switch (type) {
    case 'header':
      return `Routes based on the value of the ${param} header`
    case 'query':
      return `Routes based on the value of the ${param} query parameter`
    case 'cookie':
      return `Routes based on the value of the ${param} cookie`
    case 'user_agent':
      return `Routes based on the user agent containing ${param}`
    case 'language':
      return 'Routes based on the Accept-Language header'
    case 'expr':
      return 'Routes based on custom expressions'
    default:
      return ''
  }
}

function addExpression() {
  if (!condition.value.expressions) {
    condition.value.expressions = [];
  }
  condition.value.expressions.push({ expr: '', target: props.targets[0]?.id || '' });
  updateCondition();
}

function removeExpression(index: number) {
  if (condition.value.expressions) {
    condition.value.expressions.splice(index, 1);
    updateCondition();
  }
}
</script>
