import type { Vec4 } from '@/common/util/index.type';
import { equals } from '@/common/util/index.util';
import { inject, onUnmounted, provide, readonly, ref, watch, type Ref } from 'vue';

const STICKY_BOX_STORE_KEY = Symbol('STICKY_BOX_STORE_KEY');

export const provideStickyBoxRoot = () => {
    const ctx = {
        thisClient: ref<Vec4>([0,0,0,0]),
        rootClient: ref<Vec4>([0,0,0,0]),
        maxClient: ref<Vec4>([0,0,0,0]),
    };
    provide(STICKY_BOX_STORE_KEY, ctx)

    return ctx;
}

export const provideStickyBox = () => {

    const { thisClient, maxClient, rootClient } = useStickyBox();
    // const thieClient = ref<Vec4|null>(null);

    const ctx = {
        thisClient: ref<Vec4>([0,0,0,0]),
        rootClient: readonly(thisClient),
        maxClient
    };

    watch(thisClient, val => {
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
            rootClient.value[0],
            rootClient.value[1],
            rootClient.value[2],
            rootClient.value[3],
        ];
    })

    provide(STICKY_BOX_STORE_KEY, ctx)
    return ctx;
}

// type Context = ReturnType<typeof provideStickyBox>

type Context = {
    thisClient: Ref<Vec4>,
    rootClient: Readonly<Ref<Vec4>>
    maxClient: Ref<Vec4>
}
// export const useStickyBoxRoot = () => {
//     const root: Context = {
//     rootClient: ref<Vec4>([0,0,0,0]),
//     thisClient: ref(null),
// }
//     // try {
//     //     const { rootClient } = useStickyBox();
//     //     root = { 
//     //         rootClient,
//     //         thisClient: ref(null),
//     //     }
//     // } catch (error) {
//     //     root = {
//     //         rootClient: ref<Vec4>([0,0,0,0]),
//     //         thisClient: ref(null),
//     //     }
//     // }

//     return root as Context;
// }

export const useStickyBox = () => {
    const context = inject<Context>(STICKY_BOX_STORE_KEY)
    if (!context) throw new Error('StickyBox is not provided');
    return context;
}