import type { Vec4 } from '@/common/util/index.type';
import { equals } from '@/common/util/index.util';
import { inject, onUnmounted, provide, ref, watch, type Ref } from 'vue';

const STICKY_BOX_STORE_KEY = Symbol('STICKY_BOX_STORE_KEY');

export const provideStickyBoxRoot = () => {
    const childSlot = ref<Vec4>([0,0,0,0]);
    const rootClient = ref<Vec4>([0,0,0,0]);
    const maxClient = ref<Vec4>([0,0,0,0]);
    // 이 레벨 StickyBox들이 position:fixed 좌표 계산에 쓸 기준 사각형을, rootClient와 같은 규약으로
    // "진짜 뷰포트 각 변에서부터의 inset"([top,right,bottom,left])으로 표현. null이면 inset 없음(=진짜 뷰포트 그대로).
    // 다이얼로그처럼 조상 엘리먼트가 CSS containing block을 만들어버리는 상황(transform/contain:layout 등)에서만
    // 호출부가 그 조상의 inset으로 갱신해줘야 함 — rootClient/maxClient(조상이 예약한 공간, Math.max로 결합)와는
    // 역할이 다르므로 섞어 쓰지 말 것.
    const viewportClient = ref<Vec4 | null>(null);

    const ctx = {
        rootClient,
        maxClient,
        viewportClient,
        reportSelf: (v: Vec4) => { childSlot.value = v; },
    };

    watch(childSlot, val => {
        const newMax: Vec4 = [
            Math.max(maxClient.value[0], val[0]),
            Math.max(maxClient.value[1], val[1]),
            Math.max(maxClient.value[2], val[2]),
            Math.max(maxClient.value[3], val[3]),
        ]
        if (!equals(maxClient.value, newMax))
            maxClient.value = newMax;
    })

    provide(STICKY_BOX_STORE_KEY, ctx)

    return { rootClient, maxClient, viewportClient };
}

export const provideStickyBox = () => {

    const { rootClient: parentRootClient, maxClient, viewportClient, reportSelf: reportToParent } = useStickyBox();
    const childSlot = ref<Vec4>([0,0,0,0]);

    const ctx = {
        rootClient: parentRootClient,
        maxClient,
        viewportClient,
        reportSelf: (v: Vec4) => { childSlot.value = v; },
    };

    watch(childSlot, val => {
        const newMax: Vec4 = [
            Math.max(maxClient.value[0], val[0]),
            Math.max(maxClient.value[1], val[1]),
            Math.max(maxClient.value[2], val[2]),
            Math.max(maxClient.value[3], val[3]),
        ]
        if (!equals(maxClient.value, newMax))
            maxClient.value = newMax;
    })

    onUnmounted(() => {
        maxClient.value = [
            parentRootClient.value[0],
            parentRootClient.value[1],
            parentRootClient.value[2],
            parentRootClient.value[3],
        ];
    })

    provide(STICKY_BOX_STORE_KEY, ctx)
    return { reportSelf: reportToParent };
}

type Context = {
    rootClient: Readonly<Ref<Vec4>>
    maxClient: Ref<Vec4>
    viewportClient: Readonly<Ref<Vec4 | null>>
    reportSelf: (v: Vec4) => void
}

export const useStickyBox = () => {
    const context = inject<Context>(STICKY_BOX_STORE_KEY)
    if (!context) throw new Error('StickyBox is not provided');
    return context;
}