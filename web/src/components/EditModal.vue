<template>
  <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"></div>

  <div class="fixed inset-0 z-10 overflow-y-auto">
    <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
      <div
          class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
        <form @submit.prevent="handleSubmit">
          <div>
            <h3 class="text-base font-semibold leading-6 text-gray-900">
              {{ proxy ? 'Edit Proxy' : 'Create New Proxy' }}
            </h3>
            <div class="mt-4 space-y-4">
              <!-- Basic Information -->
              <div>
                <label for="name" class="block text-sm font-medium text-gray-700">Name</label>
                <input
                    type="text"
                    id="name"
                    v-model="form.name"
                    class="mt-1 p-2 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                    placeholder="Enter proxy name"
                />
              </div>
              
              <TagsInput v-model="form.tags" />

              <!-- Listen URLs Section -->
              <ListenUrlsSection 
                v-model="form.listen_urls" 
                :errors="urlErrors"
                @validate="validateListenUrl"
                @add="addListenUrl"
                @remove="removeListenUrl"
              />
              
              <!-- Mode Selection -->
              <div>
                <label class="block text-sm font-medium text-gray-700">Mode</label>
                <select
                    v-model="form.mode"
                    class="mt-1 block w-full p-2 rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                >
                  <option value="path">Path</option>
                  <option value="redirect">Redirect</option>
                </select>
              </div>

              <!-- Route Condition Section -->
              <RouteConditionSection 
                v-model="form.condition" 
                :targets="form.targets" 
              />

              <!-- Settings Section -->
              <SettingsSection 
                v-model="settings" 
              />

              <!-- Targets Section -->
              <TargetsSection 
                v-model="form.targets" 
              />
            </div>
          </div>
          
          <!-- Form Buttons -->
          <div class="mt-5 sm:mt-6 sm:grid sm:grid-flow-row-dense sm:grid-cols-2 sm:gap-3">
            <button
                type="submit"
                class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:col-start-2"
            >
              {{ proxy ? 'Save Changes' : 'Create' }}
            </button>
            <button
                type="button"
                @click="closeModal"
                class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:col-start-1 sm:mt-0"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import TagsInput from "@/components/TagsInput.vue";
import ListenUrlsSection from "@/components/modals/ListenUrlsSection.vue";
import TargetsSection from "@/components/modals/TargetsSection.vue";
import RouteConditionSection from "@/components/modals/RouteConditionSection.vue";
import SettingsSection from "@/components/modals/SettingsSection.vue";
import {ref, nextTick, computed} from "vue";

type Target = {
  id?: string;
  name?: string;
  url: string;
  weight: number;
  is_active?: boolean
}

type Form = {
  name: string;
  listen_url: string; // Keeping for backward compatibility
  listen_urls: string[];
  mode: string;
  tags: string[];
  targets: Target[];
  saving_cookies_flg: boolean;
  query_forwarding_flg: boolean;
  cookies_forwarding_flg: boolean;
  condition: {
    type: string;
    param_name?: string;
    values: Record<string, string>;
    expressions?: Array<{ expr: string; target: string }>;
    default?: string;
    expr?: string;
  }
}

type Proxy = {
  id: string;
}

const props = defineProps<{
  proxy: Proxy | null;
  availableTags: string[];
  form: Form;
}>()

const emit = defineEmits(['close', 'submit'])

const closeModal = () => {
  emit('close')
}

const form = ref(props.form)
const urlErrors = ref({})

// Create a computed property for the settings section that syncs with form
const settings = computed({
  get: () => ({
    saving_cookies_flg: form.value.saving_cookies_flg,
    query_forwarding_flg: form.value.query_forwarding_flg,
    cookies_forwarding_flg: form.value.cookies_forwarding_flg
  }),
  set: (val) => {
    form.value.saving_cookies_flg = val.saving_cookies_flg
    form.value.query_forwarding_flg = val.query_forwarding_flg
    form.value.cookies_forwarding_flg = val.cookies_forwarding_flg
  }
})

// Initialize listen_urls if it doesn't exist
if (!form.value.listen_urls) {
  form.value.listen_urls = form.value.listen_url ? [form.value.listen_url] : []
}

// Ensure there's at least one listen URL
if (form.value.listen_urls.length === 0) {
  form.value.listen_urls.push('')
}

const validateAllListenUrls = () => {
  // Clear all previous errors
  urlErrors.value = {}
  
  // Validate each URL
  let isValid = true
  form.value.listen_urls.forEach((url, index) => {
    // Check if URL is empty
    if (!url || url.trim() === '') {
      urlErrors.value[index] = 'Listen URL cannot be empty'
      isValid = false
      return
    }
    
    // Basic URL format validation
    const urlPattern = /^(localhost|(\d{1,3}\.){3}\d{1,3}|([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*))(:([0-9]{1,5}))?(\/.*)?$/
    if (!urlPattern.test(url)) {
      urlErrors.value[index] = 'Please enter a valid URL format'
      isValid = false
      return
    }
    
    // Check for duplicate URLs
    const duplicateIndex = form.value.listen_urls.findIndex((u, i) => u === url && i !== index)
    if (duplicateIndex !== -1) {
      urlErrors.value[index] = 'This URL is already in the list'
      isValid = false
    }
  })
  
  return isValid
}

const handleSubmit = () => {
  // Validate all listen URLs before submission
  if (!validateAllListenUrls()) {
    // Scroll to the first error
    nextTick(() => {
      const errorElement = document.querySelector('.text-red-600')
      if (errorElement) {
        errorElement.scrollIntoView({ behavior: 'smooth', block: 'center' })
      }
    })
    return
  }
  const formData = {
    name: form.value.name,
    listen_url: form.value.listen_urls[0] || '', // For backward compatibility
    listen_urls: form.value.listen_urls,
    mode: form.value.mode,
    tags: form.value.tags,
    targets: form.value.targets.map(target => ({
      ...target,
      weight: target.weight / 100
    })),
    condition: form.value.condition,
    saving_cookies_flg: form.value.saving_cookies_flg,
    query_forwarding_flg: form.value.query_forwarding_flg,
    cookies_forwarding_flg: form.value.cookies_forwarding_flg
  }

  emit('submit', formData)
}

function addListenUrl() {
  if (!form.value.listen_urls) {
    form.value.listen_urls = []
  }
  // Add a new empty URL to the list
  const newIndex = form.value.listen_urls.length
  form.value.listen_urls.push('')
  
  // Focus on the newly added input field after the DOM updates
  nextTick(() => {
    const inputs = document.querySelectorAll('input[placeholder="Enter listen URL (e.g., localhost:8080)"]')
    if (inputs && inputs[newIndex]) {
      // Cast to HTMLInputElement to access focus method
      (inputs[newIndex] as HTMLInputElement).focus()
    }
  })
}

function removeListenUrl(index) {
  form.value.listen_urls.splice(index, 1)
  // Clear any errors for this index
  delete urlErrors.value[index]
  // Reindex the errors object after removal
  const newErrors = {}
  Object.keys(urlErrors.value).forEach(key => {
    const numKey = parseInt(key)
    if (numKey > index) {
      newErrors[numKey - 1] = urlErrors.value[key]
    } else if (numKey < index) {
      newErrors[numKey] = urlErrors.value[key]
    }
  })
  urlErrors.value = newErrors
}

function validateListenUrl(index) {
  const url = form.value.listen_urls[index]
  
  // Clear previous error for this index
  delete urlErrors.value[index]
  
  // Skip validation if empty (we'll handle this during form submission)
  if (!url || url.trim() === '') {
    return
  }
  
  // Basic URL format validation
  // Allow localhost, IP addresses, and domain names with optional port
  const urlPattern = /^(localhost|((\d{1,3}\.){3}\d{1,3})|([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*))(:([0-9]{1,5}))?(\/.*)?$/
  
  if (!urlPattern.test(url)) {
    urlErrors.value[index] = 'Please enter a valid URL format'
  }
  
  // Check for duplicate URLs
  const duplicateIndex = form.value.listen_urls.findIndex((u, i) => u === url && i !== index)
  if (duplicateIndex !== -1) {
    urlErrors.value[index] = 'This URL is already in the list'
  }
}
</script>
