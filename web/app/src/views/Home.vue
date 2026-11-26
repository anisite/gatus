<template>
  <div class="dashboard-container bg-background">
    <div class="container mx-auto px-4 py-8 max-w-7xl">
      <div class="mb-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-4xl font-bold tracking-tight">{{ dashboardHeading }}</h1>
            <p class="text-muted-foreground mt-2">{{ dashboardSubheading }}</p>
          </div>
          <div class="flex items-center gap-4">
            <Button 
              variant="ghost" 
              size="icon" 
              @click="toggleShowAverageResponseTime" 
              :title="showAverageResponseTime ? $t('Show min-max response time') : $t('Show average response time')"
            >
              <Activity v-if="showAverageResponseTime" class="h-5 w-5" />
              <Timer v-else class="h-5 w-5" />
            </Button>
            <Button variant="ghost" size="icon" @click="refreshData" :title="$t('Refresh data')">
              <RefreshCw class="h-5 w-5" />
            </Button>
          </div>
        </div>
        <!-- Announcement Banner (Active Announcements) -->
        <AnnouncementBanner :announcements="activeAnnouncements" />
        <!-- Global status banner -->
        <div v-if="uiTheme === 'utd' && !loading && overallStatus !== 'unknown'" :class="['avis-utd', 'mb-4', utdAvisType]" role="status">
          <span class="avis-utd-icone">
            <CheckCircle v-if="overallStatus === 'healthy'" class="h-5 w-5" />
            <AlertCircle v-else class="h-5 w-5" />
          </span>
          <p class="avis-utd-texte">{{ overallStatusLabel }}</p>
        </div>
        <div v-else-if="!loading && overallStatus !== 'unknown'" :class="['global-status flex items-center gap-3 rounded-lg px-5 py-4 mb-4 text-white font-medium', overallStatusClass]">
          <CheckCircle v-if="overallStatus === 'healthy'" class="h-6 w-6 flex-shrink-0" />
          <AlertCircle v-else class="h-6 w-6 flex-shrink-0" />
          <span class="text-lg">{{ overallStatusLabel }}</span>
        </div>
        <!-- Search bar -->
        <SearchBar
          @search="handleSearch"
          @update:showOnlyFailing="showOnlyFailing = $event"
          @update:showRecentFailures="showRecentFailures = $event"
          @update:groupByGroup="groupByGroup = $event"
          @update:sortBy="sortBy = $event"
          @initializeCollapsedGroups="initializeCollapsedGroups"
        />
      </div>

      <div v-if="loading" class="flex items-center justify-center py-20">
        <Loading size="lg" />
      </div>

      <div v-else-if="filteredEndpoints.length === 0 && filteredSuites.length === 0" class="text-center py-20">
        <AlertCircle class="h-12 w-12 text-muted-foreground mx-auto mb-4" />
        <h3 class="text-lg font-semibold mb-2">{{ $t('No endpoints or suites found') }}</h3>
        <p class="text-muted-foreground">
          {{ searchQuery || showOnlyFailing || showRecentFailures 
            ? $t('Try adjusting your filters') 
            : $t('No endpoints or suites are configured') }}
        </p>
      </div>

      <div v-else>
        <!-- Grouped view -->
        <div v-if="groupByGroup" class="space-y-6">
          <div v-for="(items, group) in combinedGroups" :key="group" class="endpoint-group border rounded-lg overflow-hidden">
            <!-- Group Header -->
            <div 
              @click="toggleGroupCollapse(group)"
              class="endpoint-group-header flex items-center justify-between p-4 bg-card border-b cursor-pointer hover:bg-accent/50 transition-colors"
            >
              <div class="flex items-center gap-3">
                <ChevronDown v-if="uncollapsedGroups.has(group)" class="h-5 w-5 text-muted-foreground" />
                <ChevronUp v-else class="h-5 w-5 text-muted-foreground" />
                <h2 class="text-xl font-semibold text-foreground">{{ group }}</h2>
              </div>
              <div class="flex items-center gap-3">
                <span class="text-sm text-muted-foreground">{{ items.endpoints.length + items.suites.length }} services</span>
                <span v-if="calculateUnhealthyCount(items.endpoints) + calculateFailingSuitesCount(items.suites) > 0"
                      class="bg-red-600 text-white px-2 py-1 rounded-full text-sm font-medium">
                  {{ calculateUnhealthyCount(items.endpoints) + calculateFailingSuitesCount(items.suites) }}
                </span>
                <span v-else-if="calculateDegradedCount(items.endpoints) > 0"
                      class="bg-yellow-500 text-white px-2 py-1 rounded-full text-sm font-medium"
                      title="Performances dégradées">
                  {{ calculateDegradedCount(items.endpoints) }}
                </span>
                <CheckCircle v-else class="h-6 w-6 text-green-600" />
              </div>
            </div>
            
            <!-- Group Content -->
            <div v-if="uncollapsedGroups.has(group)" class="endpoint-group-content p-4">
              <!-- Suites Section -->
              <div v-if="items.suites.length > 0" class="mb-4">
                <h3 class="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3">Suites</h3>
                <div class="grid gap-3 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                  <SuiteCard
                    v-for="suite in items.suites"
                    :key="suite.key"
                    :suite="suite"
                    :maxResults="resultPageSize"
                    @showTooltip="showTooltip"
                  />
                </div>
              </div>
              
              <!-- Endpoints Section -->
              <div v-if="items.endpoints.length > 0">
                <h3 v-if="items.suites.length > 0" class="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3">Endpoints</h3>
                <div class="divide-y -mx-4 -mb-4">
                  <ServiceRow
                    v-for="endpoint in items.endpoints"
                    :key="endpoint.key"
                    :endpoint="endpoint"
                    :maxResults="resultPageSize"
                    :showAverageResponseTime="showAverageResponseTime"
                    @showTooltip="showTooltip"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Regular view -->
        <div v-else>
          <!-- Suites Section -->
          <div v-if="filteredSuites.length > 0" class="mb-6">
            <h2 class="text-lg font-semibold text-foreground mb-3">Suites</h2>
            <div class="grid gap-3 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
              <SuiteCard
                v-for="suite in paginatedSuites"
                :key="suite.key"
                :suite="suite"
                :maxResults="resultPageSize"
                @showTooltip="showTooltip"
              />
            </div>
          </div>
          
          <!-- Endpoints Section -->
          <div v-if="filteredEndpoints.length > 0">
            <h2 v-if="filteredSuites.length > 0" class="text-lg font-semibold text-foreground mb-3">Endpoints</h2>
            <div class="border rounded-lg overflow-hidden divide-y bg-card">
              <ServiceRow
                v-for="endpoint in paginatedEndpoints"
                :key="endpoint.key"
                :endpoint="endpoint"
                :maxResults="resultPageSize"
                :showAverageResponseTime="showAverageResponseTime"
                @showTooltip="showTooltip"
              />
            </div>
          </div>
        </div>

        <div v-if="!groupByGroup && totalPages > 1" class="mt-8 flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="icon"
            :disabled="currentPage === 1"
            @click="goToPage(currentPage - 1)"
          >
            <ChevronLeft class="h-4 w-4" />
          </Button>
          
          <div class="flex gap-1">
            <Button
              v-for="page in visiblePages"
              :key="page"
              :variant="page === currentPage ? 'default' : 'outline'"
              size="sm"
              @click="goToPage(page)"
            >
              {{ page }}
            </Button>
          </div>

          <Button
            variant="outline"
            size="icon"
            :disabled="currentPage === totalPages"
            @click="goToPage(currentPage + 1)"
          >
            <ChevronRight class="h-4 w-4" />
          </Button>
        </div>
      </div>

        <!-- Légende : sous la liste des endpoints -->
        <div v-if="uiTheme === 'utd' && !loading && overallStatus !== 'unknown'" class="flex flex-wrap items-center gap-4 mb-4">
          <span class="utd-pastille avec-point vert">Opérationnel (moins de 700 ms)</span>
          <span class="utd-pastille avec-point jaune">Performances dégradées (700 ms à 5 s)</span>
          <span class="utd-pastille avec-point rouge">Panne (plus de 5 s ou échec)</span>
        </div>
        <div v-else-if="!loading && overallStatus !== 'unknown'" class="flex flex-wrap items-center gap-x-5 gap-y-1 mt-6 mb-4 text-sm text-muted-foreground">
          <span class="flex items-center gap-2"><span class="inline-block w-3 h-3 rounded-sm bg-green-500"></span>Opérationnel</span>
          <span class="flex items-center gap-2"><span class="inline-block w-3 h-3 rounded-sm bg-yellow-400"></span>Performances dégradées (lenteur, pas de panne)</span>
          <span class="flex items-center gap-2"><span class="inline-block w-3 h-3 rounded-sm bg-red-500"></span>Panne</span>
        </div>
      <!-- Past Announcements Section -->
      <div v-if="archivedAnnouncements.length > 0" class="mt-12 pb-8">
        <PastAnnouncements :announcements="archivedAnnouncements" />
      </div>
    </div>

    <Settings @refreshData="fetchData" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Activity, Timer, RefreshCw, AlertCircle, ChevronLeft, ChevronRight, ChevronDown, ChevronUp, CheckCircle } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import ServiceRow from '@/components/ServiceRow.vue'
import SuiteCard from '@/components/SuiteCard.vue'
import SearchBar from '@/components/SearchBar.vue'
import Settings from '@/components/Settings.vue'
import Loading from '@/components/Loading.vue'
import AnnouncementBanner from '@/components/AnnouncementBanner.vue'
import PastAnnouncements from '@/components/PastAnnouncements.vue'
import { uiTheme } from '@/utils/uiTheme'

const props = defineProps({
  announcements: {
    type: Array,
    default: () => []
  }
})

// Computed properties for active and archived announcements
const activeAnnouncements = computed(() => {
  return props.announcements ? props.announcements.filter(a => !a.archived) : []
})

const archivedAnnouncements = computed(() => {
  return props.announcements ? props.announcements.filter(a => a.archived) : []
})

const emit = defineEmits(['showTooltip'])

const endpointStatuses = ref([])
const suiteStatuses = ref([])
const loading = ref(false)
const currentPage = ref(1)
const itemsPerPage = 96
const searchQuery = ref('')
const showOnlyFailing = ref(false)
const showRecentFailures = ref(false)
const showAverageResponseTime = ref(localStorage.getItem('gatus:show-average-response-time') !== 'false')
const groupByGroup = ref(false)
const sortBy = ref(localStorage.getItem('gatus:sort-by') || 'name')
const uncollapsedGroups = ref(new Set())
const resultPageSize = 90

const filteredEndpoints = computed(() => {
  let filtered = [...endpointStatuses.value]
  
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(endpoint => 
      endpoint.name.toLowerCase().includes(query) ||
      (endpoint.group && endpoint.group.toLowerCase().includes(query))
    )
  }
  
  if (showOnlyFailing.value) {
    filtered = filtered.filter(endpoint => {
      if (!endpoint.results || endpoint.results.length === 0) return false
      const latestResult = endpoint.results[endpoint.results.length - 1]
      return !latestResult.success
    })
  }
  
  if (showRecentFailures.value) {
    filtered = filtered.filter(endpoint => {
      if (!endpoint.results || endpoint.results.length === 0) return false
      return endpoint.results.some(result => !result.success)
    })
  }
  
  // Sort by health if selected
  if (sortBy.value === 'health') {
    filtered.sort((a, b) => {
      const aHealthy = a.results && a.results.length > 0 && a.results[a.results.length - 1].success
      const bHealthy = b.results && b.results.length > 0 && b.results[b.results.length - 1].success
      
      // Unhealthy first
      if (!aHealthy && bHealthy) return -1
      if (aHealthy && !bHealthy) return 1
      
      // Then sort by name
      return a.name.localeCompare(b.name)
    })
  }
  
  return filtered
})

const filteredSuites = computed(() => {
  let filtered = [...(suiteStatuses.value || [])]
  
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(suite => 
      suite.name.toLowerCase().includes(query) ||
      (suite.group && suite.group.toLowerCase().includes(query))
    )
  }
  
  if (showOnlyFailing.value) {
    filtered = filtered.filter(suite => {
      if (!suite.results || suite.results.length === 0) return false
      return !suite.results[suite.results.length - 1].success
    })
  }
  
  if (showRecentFailures.value) {
    filtered = filtered.filter(suite => {
      if (!suite.results || suite.results.length === 0) return false
      return suite.results.some(result => !result.success)
    })
  }
  
  // Sort by health if selected
  if (sortBy.value === 'health') {
    filtered.sort((a, b) => {
      const aHealthy = a.results && a.results.length > 0 && a.results[a.results.length - 1].success
      const bHealthy = b.results && b.results.length > 0 && b.results[b.results.length - 1].success
      
      // Unhealthy first
      if (!aHealthy && bHealthy) return -1
      if (aHealthy && !bHealthy) return 1
      
      // Then sort by name
      return a.name.localeCompare(b.name)
    })
  }
  
  return filtered
})

const totalPages = computed(() => {
  return Math.ceil((filteredEndpoints.value.length + filteredSuites.value.length) / itemsPerPage)
})

const groupedEndpoints = computed(() => {
  if (!groupByGroup.value) {
    return null
  }
  
  const grouped = {}
  filteredEndpoints.value.forEach(endpoint => {
    const group = endpoint.group || 'No Group'
    if (!grouped[group]) {
      grouped[group] = []
    }
    grouped[group].push(endpoint)
  })
  
  // Sort groups alphabetically, with 'No Group' at the end
  const sortedGroups = Object.keys(grouped).sort((a, b) => {
    if (a === 'No Group') return 1
    if (b === 'No Group') return -1
    return a.localeCompare(b)
  })
  
  const result = {}
  sortedGroups.forEach(group => {
    result[group] = grouped[group]
  })
  
  return result
})

const combinedGroups = computed(() => {
  if (!groupByGroup.value) {
    return null
  }
  
  const combined = {}
  
  // Add endpoints
  filteredEndpoints.value.forEach(endpoint => {
    const group = endpoint.group || 'No Group'
    if (!combined[group]) {
      combined[group] = { endpoints: [], suites: [] }
    }
    combined[group].endpoints.push(endpoint)
  })
  
  // Add suites
  filteredSuites.value.forEach(suite => {
    const group = suite.group || 'No Group'
    if (!combined[group]) {
      combined[group] = { endpoints: [], suites: [] }
    }
    combined[group].suites.push(suite)
  })
  
  // Sort groups alphabetically, with 'No Group' at the end
  const sortedGroups = Object.keys(combined).sort((a, b) => {
    if (a === 'No Group') return 1
    if (b === 'No Group') return -1
    return a.localeCompare(b)
  })
  
  const result = {}
  sortedGroups.forEach(group => {
    result[group] = combined[group]
  })
  
  return result
})

const paginatedEndpoints = computed(() => {
  if (groupByGroup.value) {
    // When grouping, we don't paginate
    return groupedEndpoints.value
  }
  
  const start = (currentPage.value - 1) * itemsPerPage
  const end = start + itemsPerPage
  return filteredEndpoints.value.slice(start, end)
})

const paginatedSuites = computed(() => {
  if (groupByGroup.value) {
    // When grouping, we don't paginate
    return filteredSuites.value
  }
  
  const start = (currentPage.value - 1) * itemsPerPage
  const end = start + itemsPerPage
  return filteredSuites.value.slice(start, end)
})

const visiblePages = computed(() => {
  const pages = []
  const maxVisible = 5
  let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2))
  let end = Math.min(totalPages.value, start + maxVisible - 1)
  
  if (end - start < maxVisible - 1) {
    start = Math.max(1, end - maxVisible + 1)
  }
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  
  return pages
})

const fetchData = async () => {
  // Don't show loading state on refresh to prevent UI flicker
  const isInitialLoad = endpointStatuses.value.length === 0 && suiteStatuses.value.length === 0
  if (isInitialLoad) {
    loading.value = true
  }
  try {
    // Fetch endpoints
    const endpointResponse = await fetch(`/api/v1/endpoints/statuses?page=1&pageSize=${resultPageSize}`, {
      credentials: 'include'
    })
    if (endpointResponse.status === 200) {
      const data = await endpointResponse.json()
      endpointStatuses.value = data
    } else {
      console.error('[Home][fetchData] Error fetching endpoints:', await endpointResponse.text())
    }
    
    // Fetch suites
    const suiteResponse = await fetch(`/api/v1/suites/statuses?page=1&pageSize=${resultPageSize}`, {
      credentials: 'include'
    })
    if (suiteResponse.status === 200) {
      const suiteData = await suiteResponse.json()
      suiteStatuses.value = suiteData || []
    } else {
      console.error('[Home][fetchData] Error fetching suites:', await suiteResponse.text())
      // Ensure suiteStatuses stays as empty array instead of becoming null/undefined
      if (!suiteStatuses.value) {
        suiteStatuses.value = []
      }
    }
  } catch (error) {
    console.error('[Home][fetchData] Error:', error)
  } finally {
    if (isInitialLoad) {
      loading.value = false
    }
  }
}

const refreshData = () => {
  endpointStatuses.value = [];
  suiteStatuses.value = [];
  fetchData()
}

const handleSearch = (query) => {
  searchQuery.value = query
  currentPage.value = 1
}

const goToPage = (page) => {
  currentPage.value = page
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const toggleShowAverageResponseTime = () => {
  showAverageResponseTime.value = !showAverageResponseTime.value
  localStorage.setItem('gatus:show-average-response-time', showAverageResponseTime.value ? 'true' : 'false')
}

const showTooltip = (result, event, action = 'hover') => {
  emit('showTooltip', result, event, action)
}

const calculateUnhealthyCount = (endpoints) => {
  return endpoints.filter(endpoint => {
    if (!endpoint.results || endpoint.results.length === 0) return false
    const latestResult = endpoint.results[endpoint.results.length - 1]
    return !latestResult.success
  }).length
}

const latestStatusOf = (item) => {
  if (!item.results || item.results.length === 0) return 'unknown'
  const latestResult = item.results[item.results.length - 1]
  if (!latestResult.success) return 'unhealthy'
  return latestResult.degraded ? 'degraded' : 'healthy'
}

const calculateDegradedCount = (endpoints) => {
  return endpoints.filter(endpoint => latestStatusOf(endpoint) === 'degraded').length
}

const overallStatus = computed(() => {
  const statuses = [...endpointStatuses.value, ...(suiteStatuses.value || [])].map(latestStatusOf).filter(s => s !== 'unknown')
  if (statuses.length === 0) return 'unknown'
  const unhealthy = statuses.filter(s => s === 'unhealthy').length
  if (unhealthy === statuses.length) return 'major'
  if (unhealthy > 0) return 'partial'
  if (statuses.includes('degraded')) return 'degraded'
  return 'healthy'
})

const overallStatusLabel = computed(() => {
  switch (overallStatus.value) {
    case 'major': return 'Panne majeure : aucun service ne répond'
    case 'partial': return 'Panne partielle : certains services sont en panne'
    case 'degraded': return 'Performances dégradées : certains services sont lents'
    default: return 'Tous les systèmes sont opérationnels'
  }
})

const utdAvisType = computed(() => {
  switch (overallStatus.value) {
    case 'major':
    case 'partial': return 'erreur'
    case 'degraded': return 'avertissement'
    default: return 'succes'
  }
})

const overallStatusClass = computed(() => {
  switch (overallStatus.value) {
    case 'major': return 'bg-red-600'
    case 'partial': return 'bg-orange-500'
    case 'degraded': return 'bg-yellow-500'
    default: return 'bg-green-600'
  }
})

const calculateFailingSuitesCount = (suites) => {
  return suites.filter(suite => {
    if (!suite.results || suite.results.length === 0) return false
    return !suite.results[suite.results.length - 1].success
  }).length
}

const toggleGroupCollapse = (groupName) => {
  if (uncollapsedGroups.value.has(groupName)) {
    uncollapsedGroups.value.delete(groupName)
  } else {
    uncollapsedGroups.value.add(groupName)
  }
  // Save to localStorage
  const uncollapsed = Array.from(uncollapsedGroups.value)
  localStorage.setItem('gatus:uncollapsed-groups', JSON.stringify(uncollapsed))
  localStorage.removeItem('gatus:collapsed-groups') // Remove old key if it exists
}

const initializeCollapsedGroups = () => {
  // Get saved uncollapsed groups from localStorage
  try {
    const saved = localStorage.getItem('gatus:uncollapsed-groups')
    if (saved) {
      uncollapsedGroups.value = new Set(JSON.parse(saved))
    }
    // If no saved state, uncollapsedGroups stays empty (all collapsed by default)
  } catch (e) {
    console.warn('Failed to parse saved uncollapsed groups:', e)
    localStorage.removeItem('gatus:uncollapsed-groups')
    // On error, uncollapsedGroups stays empty (all collapsed by default)
  }
}

const dashboardHeading = computed(() => {
  return window.config && window.config.dashboardHeading && window.config.dashboardHeading !== '{{ .UI.DashboardHeading }}' ? window.config.dashboardHeading : "Health Dashboard"
})

const dashboardSubheading = computed(() => {
  return window.config && window.config.dashboardSubheading && window.config.dashboardSubheading !== '{{ .UI.DashboardSubheading }}' ? window.config.dashboardSubheading : "Monitor the health of your endpoints in real-time"
})

onMounted(() => {
  fetchData()
})
</script>
