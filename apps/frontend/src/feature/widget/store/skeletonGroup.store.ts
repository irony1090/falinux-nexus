import { LifecycleRegistry } from '@/common/util/lifecycle/lifecycle.util';
import { inject, onUnmounted, provide, readonly, ref, watch } from 'vue';

const STORE_KEY = Symbol('SKELETON_GROUP_STORE_KEY');

type OffsetRect = {
    x: number
    y: number,
    width: number;
    height: number;
    radius: number;
}

export const provideSkeletonGroupStore = () => {

    const groupRef = ref<HTMLElement>();
    const isLoading = ref(false);
    const offsetRects = ref<Array<OffsetRect>>();
    // const isLoadingLazy = useLazyRef(isLoading, { startLazy: v => !v})

    const mutationManager = {
        targets: new Set<HTMLElement>(),
        observe: (el: HTMLElement) => {
            mutationManager.targets.add(el);
            mutationObserver.observe(el, { attributes: true })
        },
        unobserve: (el: HTMLElement) => {
            mutationManager.targets.delete(el);
            mutationObserver.disconnect();
            mutationManager.targets.forEach(t => mutationObserver.observe(t, { attributes: true }))
        }
    }

    const onRefresh = () => {
        // const rects: Array<DOMRect> = [];
        if (!groupRef.value || !isLoading.value) {
            offsetRects.value = undefined
            return
        }
        const { x: gx, y: gy } = groupRef.value.getBoundingClientRect();
        const newRects: Array<OffsetRect> = [];
        for (const el of mutationManager.targets.values()) {
            const { x, y, width, height } = el.getBoundingClientRect()
            newRects.push({
                x: x - gx,
                y: y - gy,
                width,
                height,
                radius: Math.min(width, height) / 2,
            })
        }

        offsetRects.value = newRects;
        
    }
    
    const resizeObserver = new ResizeObserver(onRefresh);
    const mutationObserver = new MutationObserver(onRefresh);

    const lifecycleRegistry = new LifecycleRegistry<HTMLElement>({
        init: el => {
            // console.log('[LIFE_CYCLE] INIT', el)
            resizeObserver.observe(el);
            mutationManager.observe(el);
        },
        release: el =>{
            // console.log('[LIFE_CYCLE] RELEASE', el)
            resizeObserver.unobserve(el);
            mutationManager.unobserve(el);
            onRefresh();
        },
    })
    lifecycleRegistry.on('registered', _ => {
        if (isLoading.value) lifecycleRegistry.initAll();
    })

    const registryElement = lifecycleRegistry.put.bind(lifecycleRegistry);
    const releaseElement = lifecycleRegistry.remove.bind(lifecycleRegistry);

    watch(isLoading, val => {
        if (val) lifecycleRegistry.initAll();
        else lifecycleRegistry.releaseAll()
        
    }, { immediate: true })


    const context = { 
        isLoading,
        groupRef,
        offsetRects,
        registryElement,
        releaseElement,
     }

    provide(STORE_KEY, context)
    return context
}

type Context = ReturnType<typeof provideSkeletonGroupStore>

export const useSkeletonGroupStore = (): Context => {
    const context = inject<Context>(STORE_KEY, { 
        isLoading: readonly(ref(false)),
        registryElement: _ => {},
        releaseElement: _ => {},
        groupRef: readonly(ref(undefined)),
        offsetRects: readonly(ref(undefined)),
    });

    return context;
}