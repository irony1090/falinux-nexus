<script setup lang="ts">
import { computeResizeSize, useElementsChange, type UseResizeSizeValue } from '@/common/hook/vue.hook';
import { computed, nextTick, onMounted, onUnmounted, type CSSProperties, type PropType, ref, useSlots, watch } from 'vue';
import { provideStickyBox, useStickyBox } from '../store/stickyBox.store';
import type { Vec4 } from '@/common/util/index.type';
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';

type RelationTarget = HTMLElement | null | undefined;

const { rootClient, viewportClient, reportSelf: reportToParent } = useStickyBox();
const props = defineProps({
    name: {
        type: String,
    },
    // 이 StickyBox 밖에서 일어나는, 위치/크기 재계산이 필요한 변화의 원천(들).
    // scroll/resize/속성변화를 모두 감지한다 (예: 다이얼로그 카드, document.documentElement 등).
    relation: {
        type: [Object, Array] as PropType<RelationTarget | RelationTarget[]>,
        default: () => null,
    }
})

const relations = computed<RelationTarget[]>(() =>
    Array.isArray(props.relation) ? props.relation : [props.relation]
)

// const stickyResize = useResize(stickyRef);
const stickyRef = ref<HTMLElement|null>(null);
const stickyResize = ref<UseResizeSizeValue|null>(null)
useResizeCallback(stickyRef, () => {
    stickyResize.value = computeResizeSize(stickyRef.value);
})

const thisRectClient = ref<Vec4>([0, 0, 0, 0]);

const slots = useSlots();

const headRef = ref<HTMLElement|null>(null);
const footRef = ref<HTMLElement|null>(null);

// const hSize = useResize(headRef);
const hSize = ref<UseResizeSizeValue|null>(null)
useResizeCallback(headRef, () => {
    hSize.value = computeResizeSize(headRef.value);
})
// const fSize = useResize(footRef);
const fSize = ref<UseResizeSizeValue|null>(null)
useResizeCallback(footRef, () => {
    fSize.value = computeResizeSize(footRef.value);
})

const headStyle = computed<CSSProperties>(() => {
    return {
        top: `${thisRectClient.value[0]}px`,
        right: `${thisRectClient.value[1]}px`,
        left: `${thisRectClient.value[3]}px`,
        '--sticky-height': stickyResize.value?.padding[2] !== undefined ? `${stickyResize.value.padding[2]}px`  : 'auto'
    }
})
const footStyle = computed<CSSProperties>(() => {
    return {
        right: `${thisRectClient.value[1]}px`,
        bottom: `${thisRectClient.value[2]}px`,
        left: `${thisRectClient.value[3]}px`,
        '--sticky-height': stickyResize.value?.padding[2] !== undefined ? `${stickyResize.value.padding[2]}px`  : 'auto'
    }
})

const boxStyle = computed<CSSProperties>(() => {
    return {
        paddingTop: `${(hSize.value?.outer[1] ?? 0)}px`,
        paddingBottom: `${(fSize.value?.outer[1] ?? 0)}px`,
    }
})

const refresh = () => {
    if (!stickyRef.value) {
        reportToParent([0, 0, 0, 0]);
        return;
    }
    const [rootT, rootR, rootB, rootL] = rootClient.value
    // viewportClient는 rootClient와 같은 규약(진짜 뷰포트 각 변에서부터의 inset). 없으면 inset 0(=진짜 뷰포트 그대로)
    // (다이얼로그처럼 조상이 CSS containing block을 만드는 경우 store 쪽에서 그 조상의 inset으로 보정해줌).
    const [, vRight, vBottom, vLeft] = viewportClient.value ?? [0, 0, 0, 0];
    const { clientHeight, clientWidth } = document.documentElement;
    const { bottom, left, width } = stickyRef.value.getBoundingClientRect()
    const clientB = (clientHeight - bottom) - vBottom;
    thisRectClient.value = [
        Math.max(0, rootT),
        Math.max(0, (clientWidth - (width + left)) - vRight, rootR),
        Math.max(0, clientB, rootB),
        Math.max(0, (left - vLeft), rootL),
    ]
}

watch([
    thisRectClient,
    () => (hSize.value?.rect.height ?? 0),
    () => fSize.value?.rect.height ?? 0
], ([rectC, hh, fh]) => {
    reportToParent([
        rectC[0] + hh,
        rectC[1],
        rectC[2] + fh,
        rectC[3]
    ])
})

// 자기 자신의 크기 변화(useResize) + 부모가 알려주는 root clamp/기준 사각형 변화
watch([stickyResize, rootClient, viewportClient], refresh)
// head/foot 자신의 실측 높이가 확정되는 시점(=boxStyle의 padding이 반영되는 시점)에 맞춰 재계산.
// stickyResize의 재관찰(리사이즈옵저버가 padding 반영 후 다시 감지)에만 의존하면 타이밍이 어긋날 수 있어 명시적으로 보강.
watch([
    () => hSize.value?.outer[1],
    () => fSize.value?.outer[1],
], () => nextTick(refresh))
// relation 자체가 교체될 때(예: null -> 실제 엘리먼트) 즉시 한번 재계산
watch(relations, refresh)
// relation 대상들의 scroll/resize/속성변화
useElementsChange(relations, refresh)

// 뷰포트 자체가 리사이즈되는 경우도 clamp 재계산 대상
onMounted(() => window.addEventListener('resize', refresh))
onUnmounted(() => window.removeEventListener('resize', refresh))

defineExpose({
    stickyResize,
    rectClient: thisRectClient
})

provideStickyBox();
</script>

<template>
<div class="stickyBox" ref="stickyRef" :style="boxStyle" >
    <div v-if="slots.header" ref="headRef" class="head" :style="headStyle">
        <slot name="header" />
    </div>
    <slot default/>
    <div v-if="slots.footer" ref="footRef" class="foot" :style="footStyle">
        <slot name="footer" />
    </div>
</div>
</template>
<style lang="scss" scoped>
.stickyBox {
    overflow: hidden;
    .head, .foot {
        z-index: 1;
        position: fixed;
        left: 0;
        right: 0;
    }
}
</style>
