<script setup>
import { ref, onMounted } from 'vue'
import { 
  ShoppingCart, Search, Tag, ArrowRight, 
  Dumbbell, Utensils, Zap, Shirt
} from 'lucide-vue-next'
import { getProducts } from '../api'

const props = defineProps({
  membershipType: String
})

const loading = ref(false)
const products = ref([])
const activeCategory = ref('全部')

const categories = [
  { name: '全部', icon: ShoppingCart },
  { name: '健身器材', icon: Dumbbell },
  { name: '营养补剂', icon: Utensils },
  { name: '运动装备', icon: Shirt },
  { name: '健康食品', icon: Zap }
]

const fetchProducts = async (category) => {
  loading.value = true
  activeCategory.value = category
  try {
    const res = await getProducts(category === '全部' ? '' : category)
    products.value = res.data || res // Handle both Axios and direct return
  } catch (err) {
    console.error('Fetch products failed:', err)
  } finally {
    loading.value = false
  }
}

const getDiscountPrice = (price) => {
  if (props.membershipType && props.membershipType !== 'FREE') {
    return (price * 0.95).toFixed(2)
  }
  return null
}

onMounted(() => fetchProducts('全部'))
</script>

<template>
  <div class="h-full flex flex-col bg-background p-6 space-y-8">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-6">
      <div class="space-y-1">
        <h2 class="text-3xl font-black tracking-tight text-foreground flex items-center gap-3">
          <ShoppingCart class="w-8 h-8 text-orange-500" />
          健身商城
        </h2>
        <p class="text-muted-foreground font-medium">挑选您的健康伙伴，助力每一次蜕变</p>
      </div>
      
      <div class="relative max-w-sm w-full">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        <input 
          type="text" 
          placeholder="搜索商品..." 
          class="w-full bg-muted/50 border border-border rounded-2xl py-3 pl-12 pr-4 focus:border-primary transition-all outline-none text-sm"
        />
      </div>
    </div>

    <!-- Category Tabs -->
    <div class="flex flex-wrap gap-2">
      <button 
        v-for="cat in categories" 
        :key="cat.name"
        @click="fetchProducts(cat.name)"
        class="flex items-center gap-2 px-6 py-2.5 rounded-2xl font-bold text-sm transition-all whitespace-nowrap"
        :class="[
          activeCategory === cat.name 
            ? 'bg-primary text-primary-foreground shadow-lg shadow-primary/20' 
            : 'bg-muted/50 text-muted-foreground hover:bg-muted border border-transparent'
        ]"
      >
        <component :is="cat.icon" class="w-4 h-4" />
        {{ cat.name }}
      </button>
    </div>

    <!-- Membership Perk Banner -->
    <div v-if="membershipType !== 'FREE'" class="p-4 rounded-2xl bg-orange-500/10 border border-orange-500/20 flex items-center justify-between gap-4">
      <div class="flex items-center gap-3 text-orange-600 dark:text-orange-400">
        <Tag class="w-5 h-5" />
        <span class="font-bold text-sm">PRO会员专享：全场商品 95 折优惠已生效</span>
      </div>
      <ArrowRight class="w-5 h-5 text-orange-500" />
    </div>

    <!-- Product Grid -->
    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <div class="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
    </div>
    <div v-else-if="products.length === 0" class="flex-1 flex flex-col items-center justify-center text-muted-foreground space-y-4">
      <ShoppingCart class="w-16 h-16 opacity-20" />
      <p class="font-bold">暂无相关商品</p>
    </div>
    <div v-else class="grid grid-cols-2 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 md:gap-6 pb-12">
      <div 
        v-for="item in products" 
        :key="item.id"
        class="group bg-card border border-border rounded-2xl md:rounded-3xl overflow-hidden hover:shadow-xl hover:-translate-y-1 transition-all duration-300 flex flex-col"
      >
        <div class="aspect-square bg-muted flex items-center justify-center relative flex-shrink-0">
          <!-- Placeholder for actual image -->
          <div class="text-xl md:text-4xl font-black text-muted-foreground/20 uppercase tracking-tighter text-center px-2 md:px-4">
            {{ item.name }}
          </div>
          <div v-if="getDiscountPrice(item.price)" class="absolute top-2 left-2 md:top-4 md:left-4 px-2 py-0.5 md:px-3 md:py-1 bg-orange-500 text-white text-[8px] md:text-[10px] font-black rounded-full uppercase tracking-widest shadow-lg">
            9.5折
          </div>
        </div>
        
        <div class="p-3 md:p-6 space-y-2 md:space-y-4 flex-1 flex flex-col justify-between">
          <div class="space-y-0.5 md:space-y-1">
            <div class="text-[8px] md:text-xs font-bold text-primary uppercase tracking-widest">{{ item.category }}</div>
            <h3 class="text-xs md:text-base font-bold text-foreground line-clamp-1">{{ item.name }}</h3>
          </div>

          <div class="flex items-end justify-between gap-1">
            <div class="space-y-0 md:space-y-0.5 min-w-0">
              <div v-if="getDiscountPrice(item.price)" class="text-[8px] md:text-xs text-muted-foreground line-through">¥{{ item.price }}</div>
              <div class="text-sm md:text-xl font-black text-foreground truncate">¥{{ getDiscountPrice(item.price) || item.price }}</div>
            </div>
            <button class="p-2 md:p-3 bg-primary text-primary-foreground rounded-xl md:rounded-2xl hover:scale-110 transition-transform shadow-lg shadow-primary/20 shrink-0">
              <ShoppingCart class="w-3 h-3 md:w-5 md:h-5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
