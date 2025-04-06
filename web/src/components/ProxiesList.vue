<template>
  <div class="mt-8 flow-root">
    <div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
      <div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
        <div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
          <table class="min-w-full divide-y divide-gray-300">
            <thead class="bg-gray-50">
            <tr>
              <th
                  scope="col"
                  class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6 cursor-pointer"
                  @click="handleSort('id')"
              >
                ID
                <span v-if="sortBy === 'id'" class="ml-2 flex-none rounded text-gray-400 group-hover:visible group-focus:visible">
                  <ChevronDownIcon v-if="sortDesc" class="h-5 w-5" aria-hidden="true"/>
                  <ChevronUpIcon v-if="!sortDesc" class="h-5 w-5" aria-hidden="true"/>
                </span>
              </th>
              <th
                  scope="col"
                  class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 cursor-pointer"
                  @click="handleSort('name')"
              >
                Name
                <span v-if="sortBy === 'name'" class="ml-2 flex-none rounded text-gray-400 group-hover:visible group-focus:visible">
                  <ChevronDownIcon v-if="sortDesc" class="h-5 w-5" aria-hidden="true"/>
                  <ChevronUpIcon v-if="!sortDesc" class="h-5 w-5" aria-hidden="true"/>
                </span>
              </th>
              <th
                  scope="col"
                  class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 cursor-pointer"
                  @click="handleSort('mode')"
              >
                Mode
                <span v-if="sortBy === 'mode'" class="ml-2 flex-none rounded text-gray-400 group-hover:visible group-focus:visible">
                  <ChevronDownIcon v-if="sortDesc" class="h-5 w-5" aria-hidden="true"/>
                  <ChevronUpIcon v-if="!sortDesc" class="h-5 w-5" aria-hidden="true"/>
                </span>
              </th>
              <th
                  scope="col"
                  class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 cursor-pointer"
                  @click="handleSort('listen_url')"
              >
                Listen URL
                <span v-if="sortBy === 'listen_url'" class="ml-2 flex-none rounded text-gray-400 group-hover:visible group-focus:visible">
                  <ChevronDownIcon v-if="sortDesc" class="h-5 w-5" aria-hidden="true"/>
                  <ChevronUpIcon v-if="!sortDesc" class="h-5 w-5" aria-hidden="true"/>
                </span>
              </th>
              <th
                  scope="col"
                  class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 cursor-pointer"
                  @click="handleSort('targets')"
              >
                Targets
                <span v-if="sortBy === 'targets'" class="ml-2 flex-none rounded text-gray-400 group-hover:visible group-focus:visible">
                  <ChevronDownIcon v-if="sortDesc" class="h-5 w-5" aria-hidden="true"/>
                  <ChevronUpIcon v-if="!sortDesc" class="h-5 w-5" aria-hidden="true"/>
                </span>
              </th>
              <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Tags</th>
              <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Cookies</th>
              <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Forwarding</th>
              <th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-6">
                <span class="sr-only">Actions</span>
              </th>
            </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 bg-white">
            <tr v-for="proxy in filteredProxies" :key="proxy.id">
              <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                <span class="cursor-pointer hover:text-indigo-600" @click="editProxy(proxy)">{{ `${proxy.id.slice(0, 4)}...${proxy.id.slice(-4)}` }}</span>
              </td>
              <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                {{ proxy.name || 'Unnamed Proxy' }}
              </td>
              <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ proxy.mode }}</td>
              <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                <div v-for="url in proxy.listen_urls" :key="url.id">
                  {{ url.listen_url }}
                  <span v-if="url.path_key" class="text-xs text-gray-400">({{ url.path_key }})</span>
                </div>
              </td>
              <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                <div v-for="target in proxy.targets" :key="target.id">
                  {{ target.url }} ({{ (target.weight * 100).toFixed(0) }}%)
                  <span v-if="!target.is_active" class="text-xs text-red-500">(inactive)</span>
                </div>
              </td>
              <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                <div class="flex flex-wrap gap-1">
                  <span
                      v-for="tag in proxy.tags"
                      :key="tag"
                      class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-800"
                  >
                    {{ tag }}
                  </span>
                </div>
              </td>
              <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                <span :class="proxy.saving_cookies_flg ? 'text-green-600' : 'text-gray-400'">
                  {{ proxy.saving_cookies_flg ? 'Enabled' : 'Disabled' }}
                </span>
              </td>
              <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                <span v-if="proxy.query_forwarding_flg" class="text-green-600">Query</span>
                <span v-if="proxy.cookies_forwarding_flg" class="text-green-600">Cookies</span>
                <span v-if="!proxy.query_forwarding_flg && !proxy.cookies_forwarding_flg" class="text-gray-400">
                  <NoSymbolIcon class="h-5 w-5" aria-hidden="true"/>
                </span>
              </td>
              <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                <button
                    @click="viewHistory(proxy)"
                    class="text-indigo-600 hover:text-indigo-900 mr-2"
                >
                  <ArchiveBoxIcon class="h-5 w-5" aria-hidden="true"/>
                </button>
                <button
                    @click="deleteProxy(proxy.id)"
                    class="text-red-600 hover:text-red-900"
                >
                  <TrashIcon class="h-5 w-5" aria-hidden="true"/>
                </button>
              </td>
            </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ChevronDownIcon, ChevronUpIcon, TrashIcon, ArchiveBoxIcon, NoSymbolIcon} from '@heroicons/vue/20/solid'

type Proxy = {
  id: string,
  listen_url?: string,
  path_key?: string,
  mode: string,
  tags: string[],
  targets: Array<{ id: string, url: string, weight: number, is_active: boolean }>,
  name?: string,
  saving_cookies_flg: boolean,
  query_forwarding_flg: boolean,
  cookies_forwarding_flg: boolean,
  listen_urls: Array<{ id: string, listen_url: string, path_key?: string }>
}

const props = defineProps<{
  filteredProxies: Array<Proxy>,
  sortBy: string,
  sortDesc: boolean
}>()

const emit = defineEmits(['delete', 'edit', 'viewHistory', 'sort'])

const handleSort = (column) => {
  emit('sort', column)
}

const editProxy = (proxy) => {
  emit('edit', proxy)
}

const viewHistory = (proxy) => {
  emit('viewHistory', proxy)
}

const deleteProxy = (id) => {
  emit('delete', id)
}
</script>