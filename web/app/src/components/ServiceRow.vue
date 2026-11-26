<template>
  <div class="service-row px-4 sm:px-6 py-2.5">
    <!-- Status line: icon, name, status label, expand button -->
    <div class="flex flex-wrap items-center justify-between gap-x-3 gap-y-2">
      <!-- Le titre mène à la page de l'endpoint ; seul le + / - ouvre les infos. Le nom passe à la ligne au besoin (mobile). -->
      <router-link :to="`/endpoints/${endpoint.key}`" class="flex items-start sm:items-center gap-3 min-w-0 flex-1 basis-[16rem] hover:underline">
        <CheckCircle v-if="currentStatus === 'healthy'" class="h-5 w-5 flex-shrink-0 text-emerald-500" />
        <AlertTriangle v-else-if="currentStatus === 'degraded'" class="h-5 w-5 flex-shrink-0 text-yellow-500" />
        <XCircle v-else-if="currentStatus === 'unhealthy'" class="h-5 w-5 flex-shrink-0 text-red-500" />
        <HelpCircle v-else class="h-5 w-5 flex-shrink-0 text-muted-foreground" />
        <span class="font-medium break-words min-w-0" :title="endpoint.name">
          {{ endpoint.name }}
          <!-- Le groupe distingue les endpoints qui portent le même nom (ex. FRW-IT et FRW-PROD) -->
          <span v-if="endpoint.group" class="service-groupe ml-2 align-middle text-xs font-normal text-muted-foreground border rounded px-1.5 py-0.5">{{ endpoint.group }}</span>
        </span>
      </router-link>
      <div class="flex items-center gap-3 flex-shrink-0 text-sm ml-auto">
        <template v-if="uiTheme === 'utd'">
          <span class="text-muted-foreground">{{ formattedResponseTime }}</span>
          <span :class="['utd-pastille', 'avec-point', utdPastilleColor]">{{ statusLabel }}</span>
        </template>
        <span v-else :class="statusTextClass">{{ statusLabel }}</span>
        <button
          type="button"
          class="p-1 rounded hover:bg-accent"
          :aria-expanded="expanded"
          :aria-label="expanded ? 'Masquer les détails' : 'Afficher les détails'"
          @click="expanded = !expanded"
        >
          <Minus v-if="expanded" class="h-4 w-4 text-muted-foreground" />
          <Plus v-else class="h-4 w-4 text-muted-foreground" />
        </button>
      </div>
    </div>

    <!-- History bar -->
    <div class="historique-barres flex gap-[2px] mt-2">
      <div
        v-for="(result, index) in displayResults"
        :key="index"
        :class="[
          'flex-1 h-5 sm:h-6 rounded-[2px] transition-colors',
          result ? 'cursor-pointer' : '',
          segmentClass(result, index)
        ]"
        @mouseenter="result && handleMouseEnter(result, $event)"
        @mouseleave="result && handleMouseLeave(result, $event)"
        @click.stop="result && handleClick(result, $event, index)"
      />
    </div>

    <!-- Details -->
    <div v-if="expanded" class="mt-4 text-sm space-y-3">
      <div class="flex flex-wrap gap-x-8 gap-y-1 text-muted-foreground">
        <span v-if="hostname">{{ hostname }}</span>
        <span>Temps de réponse : {{ formattedResponseTime }}</span>
        <span v-if="newestResultTime">Dernière vérification : {{ newestResultTime }}</span>
      </div>
      <div v-if="latestResult && latestResult.conditionResults && latestResult.conditionResults.length" class="font-mono text-xs space-y-0.5">
        <div v-for="(conditionResult, index) in latestResult.conditionResults" :key="index" class="flex items-start gap-2">
          <span :class="conditionResult.success ? 'text-emerald-500' : (latestResult.degraded ? 'text-yellow-500' : 'text-red-500')">
            {{ conditionResult.success ? '✓' : (latestResult.degraded ? '!' : '✗') }}
          </span>
          <span class="break-all">{{ conditionResult.condition }}</span>
        </div>
      </div>
      <div v-if="latestResult && latestResult.errors && latestResult.errors.length" class="font-mono text-xs text-red-500">
        <div v-for="(error, index) in latestResult.errors" :key="index">• {{ error }}</div>
      </div>
      <button class="text-xs text-primary hover:underline" @click.stop="navigateToDetails">Voir les détails</button>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { CheckCircle, AlertTriangle, XCircle, HelpCircle, Plus, Minus } from 'lucide-vue-next'
import { generatePrettyTimeAgo } from '@/utils/time'
import { uiTheme } from '@/utils/uiTheme'

const router = useRouter()

const props = defineProps({
  endpoint: {
    type: Object,
    required: true
  },
  maxResults: {
    type: Number,
    default: 90
  },
  showAverageResponseTime: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['showTooltip'])

const expanded = ref(false)
const selectedResultIndex = ref(null)

const latestResult = computed(() => {
  if (!props.endpoint.results || props.endpoint.results.length === 0) {
    return null
  }
  return props.endpoint.results[props.endpoint.results.length - 1]
})

const currentStatus = computed(() => {
  if (!latestResult.value) return 'unknown'
  if (latestResult.value.degraded) return 'degraded'
  return latestResult.value.success ? 'healthy' : 'unhealthy'
})

const statusLabel = computed(() => {
  switch (currentStatus.value) {
    case 'healthy': return 'Opérationnel'
    case 'degraded': return 'Performances dégradées'
    case 'unhealthy': return 'Panne'
    default: return 'Inconnu'
  }
})

const utdPastilleColor = computed(() => {
  switch (currentStatus.value) {
    case 'healthy': return 'vert'
    case 'degraded': return 'jaune'
    case 'unhealthy': return 'rouge'
    default: return 'bleu'
  }
})

const statusTextClass = computed(() => {
  switch (currentStatus.value) {
    case 'degraded': return 'text-yellow-600 dark:text-yellow-400'
    case 'unhealthy': return 'text-red-600 dark:text-red-400'
    default: return 'text-muted-foreground'
  }
})

const hostname = computed(() => latestResult.value?.hostname || null)

// Nombre de carrés selon la largeur de l'écran : 90 est trop sur mobile
const viewportWidth = ref(window.innerWidth)
const updateViewportWidth = () => { viewportWidth.value = window.innerWidth }
const visibleCount = computed(() => {
  if (viewportWidth.value < 640) return Math.min(30, props.maxResults)
  if (viewportWidth.value < 1024) return Math.min(60, props.maxResults)
  return props.maxResults
})

const displayResults = computed(() => {
  const results = [...(props.endpoint.results || [])]
  while (results.length < visibleCount.value) {
    results.unshift(null)
  }
  return results.slice(-visibleCount.value)
})

const segmentClass = (result, index) => {
  const selected = selectedResultIndex.value === index
  if (!result) return 'bg-gray-200 dark:bg-gray-700'
  if (result.degraded) return selected ? 'bg-yellow-600' : 'bg-yellow-400 hover:bg-yellow-600'
  if (result.success) return selected ? 'bg-emerald-700' : 'bg-emerald-500 hover:bg-emerald-700'
  return selected ? 'bg-red-700' : 'bg-red-500 hover:bg-red-700'
}

const formattedResponseTime = computed(() => {
  const durations = (props.endpoint.results || []).filter(r => r.duration).map(r => r.duration / 1000000)
  if (durations.length === 0) return 'N/A'
  if (props.showAverageResponseTime) {
    return `~${Math.round(durations.reduce((a, b) => a + b, 0) / durations.length)}ms`
  }
  const min = Math.trunc(Math.min(...durations))
  const max = Math.trunc(Math.max(...durations))
  return min === max ? `${min}ms` : `${min}-${max}ms`
})

const newestResultTime = computed(() => {
  if (!latestResult.value) return ''
  return generatePrettyTimeAgo(latestResult.value.timestamp)
})

const navigateToDetails = () => {
  router.push(`/endpoints/${props.endpoint.key}`)
}

const handleMouseEnter = (result, event) => {
  emit('showTooltip', result, event, 'hover')
}

const handleMouseLeave = (result, event) => {
  emit('showTooltip', null, event, 'hover')
}

const handleClick = (result, event, index) => {
  // Clear selections in other rows first
  window.dispatchEvent(new CustomEvent('clear-data-point-selection'))
  if (selectedResultIndex.value === index) {
    selectedResultIndex.value = null
    emit('showTooltip', null, event, 'click')
  } else {
    selectedResultIndex.value = index
    emit('showTooltip', result, event, 'click')
  }
}

const handleClearSelection = () => {
  selectedResultIndex.value = null
}

onMounted(() => {
  window.addEventListener('clear-data-point-selection', handleClearSelection)
  window.addEventListener('resize', updateViewportWidth)
})

onUnmounted(() => {
  window.removeEventListener('clear-data-point-selection', handleClearSelection)
  window.removeEventListener('resize', updateViewportWidth)
})
</script>
