<script setup lang="ts">
import { provideSkeletonGroupStore } from '@/feature/widget/store/skeletonGroup.store';
import { computed, toRefs, watch, type CSSProperties, type PropType } from 'vue';

const { isLoading, groupRef, offsetRects } = provideSkeletonGroupStore();

const props = defineProps({
    colors: {
        type: Object as PropType<Record<string, string>>,
        default: () => ({})
    },
    bgColor: {
        type: String,
        default: () => 'transparent'
    },
    highlightColor: {
        type: String,
        default: () => 'white'
    },
    color: {
        type: String,
        default: () => 'grey'
    },
})
const { bgColor: bgC, highlightColor: hlColor, color: c, colors } = toRefs(props);

const bg = computed(() => {
    return colors.value[bgC.value] ?? bgC.value
})
const highlight = computed(() => {
    return colors.value[hlColor.value] ?? hlColor.value
})
// const color = computed(() => colors.value[c.value] ?? c.value);

const isLoadingModel = defineModel<boolean>({
    type: Boolean,
    default: () => false,
})

const cls = computed(() => ({loading: isLoadingModel.value}))

watch(isLoadingModel, val => {
    isLoading.value = val;
}, { immediate: true })

const skeletonGroupStyle = computed<CSSProperties>(() => ({
    '--skeleton-color': colors.value[c.value] ?? c.value
}))

// defineExpose(context)
</script>
<template>
<div class="SkeletonGroup"
    ref="groupRef"
    :class="cls"
    :style="skeletonGroupStyle"
>
    <slot />
    <div class="shimmer-mask"></div>
    <svg v-if="offsetRects"
        width="100%" height="100%" style="position: absolute; top:0; left:0;"
    >
        <clipPath id="CLIP_PATH" >
            <rect v-for="{ x, y, width, height, radius }, i in offsetRects"
                :key="i"
                :x="`${x}px`"
                :y="`${y}px`"
                :width="`${width}px`"
                :height="`${height}px`"
                :rx="radius"
                :ry="radius"
            />
        </clipPath>
    </svg>
</div>
</template>

<style scoped lang="scss">
$--bg: v-bind(bg);
$--highlight: v-bind(highlight);
//$--skeleton-color: v-bind(color);

@keyframes loading {
    100% {
        transform: translate(100%);
    }
}

.SkeletonGroup {
    position: relative;
    .shimmer-mask {
    //&::after {
        display: none;
    }
    &.loading {
        background-color: $--bg;
        .shimmer-mask {
        //&::after {
            clip-path: url(#CLIP_PATH);
            //clip-path: polygon(50% 0%, 0% 100%, 100% 100%);
            display: initial;
            z-index: 1;
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            &::after {
                content: "";
                position: absolute;
                top: 0; left: 0;
                width: 100%; height: 100%;
                background: linear-gradient(90deg, 
                    color-mix(in srgb, transparent 10%, transparent), 
                    color-mix(in srgb, $--highlight 5%, transparent), 
                    color-mix(in srgb, $--highlight 45%, transparent), 
                    color-mix(in srgb, $--highlight 5%, transparent),
                    color-mix(in srgb, transparent 10%, transparent) 
                );
                contain: content;
                will-change: transform;
                animation: loading 1.5s infinite;
                transform: translate(-100%);
            }
        }
    }
    
}
</style>