<template>
  <table class="w-full text-sm text-left border-collapse bg-card rounded-xl overflow-hidden">
    <thead class="bg-muted text-muted-foreground uppercase text-xs">
      <tr>
        <th class="px-4 py-3 border-b border-border">项目 (每100g)</th>
        <th class="px-4 py-3 border-b border-border">数值</th>
      </tr>
    </thead>
    <tbody class="divide-y divide-border">
      <tr>
        <td class="px-4 py-3 font-medium text-muted-foreground">食物名称</td>
        <td class="px-4 py-3 text-foreground">
          {{ data.food_name_zh || data.food_name }}
          <span v-if="data.food_name_en" class="text-xs text-muted-foreground block">({{ data.food_name_en }})</span>
        </td>
      </tr>
      <tr>
        <td class="px-4 py-3 font-medium text-muted-foreground">能量 (Calories)</td>
        <td class="px-4 py-3 text-foreground">{{ data.calories_per_100g }} kcal</td>
      </tr>
      <tr>
        <td class="px-4 py-3 font-medium text-muted-foreground">蛋白质 (Protein)</td>
        <td class="px-4 py-3 text-foreground">{{ data.protein_per_100g }} g</td>
      </tr>
      <tr>
        <td class="px-4 py-3 font-medium text-muted-foreground">脂肪 (Fat)</td>
        <td class="px-4 py-3 text-foreground">{{ data.fat_per_100g }} g</td>
      </tr>
      <tr>
        <td class="px-4 py-3 font-medium text-muted-foreground">碳水化合物 (Carbs)</td>
        <td class="px-4 py-3 text-foreground">{{ data.carbs_per_100g }} g</td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
defineProps({
  data: {
    type: Object,
    required: true
  }
})
</script>
