<template>
<component 
    class="Skeleton"
    :is="component" 
    :class="{loading}" 
    :style="{minWidth: loading ? minWidth : undefined}"
    ref="skeletonRef"
    v-bind="$attrs" 
>
    <slot v-if="!slots.skeleton || !loading"></slot>
    <slot v-else name="skeleton"></slot>
</component>
</template>

<script lang="ts" setup>
import { useSkeletonGroupStore } from '@/feature/widget/store/skeletonGroup.store';
import { computed, onBeforeUnmount, ref, toRefs, useSlots, watch } from 'vue';

const props = defineProps({
    component: {
        type: String,
        default: () => 'div'
    },
    minWidth: {
        type: String,
        default: () => null,
    },
    loading: {
        type: Boolean,
        default: () => false
    },
    color: {
        type: String,
        default: () => 'grey'
    },
})

const { component, loading: pLoading, minWidth, color } = toRefs(props);
const { isLoading, registryElement, releaseElement } = useSkeletonGroupStore();

const slots = useSlots();
const skeletonRef = ref<HTMLElement>()

watch(skeletonRef, (val, oldVal) => {
    if (val) {
        registryElement(val)
    } else if (oldVal) {
        releaseElement(oldVal)
    }
})

onBeforeUnmount(() => {
    if (skeletonRef.value) releaseElement(skeletonRef.value);
})

const loading = computed(() => pLoading.value || isLoading.value);
</script>

<style scoped lang="scss">
//$--skeleton-color: var(--skeleton-color, v-bind(color));

.Skeleton {
    &.loading {
        position: relative;
        border-radius: 999px;
        overflow: hidden;
        opacity: 0.625;
        &::after {
            content: '';
            position: absolute;
            width: 100%;
            height: 100%;
            left: 0;
            top: 0;
            background-color: var(--skeleton-color, v-bind(color));
        }
    }
}
</style>