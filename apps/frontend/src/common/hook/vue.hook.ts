import { onUnmounted, readonly, ref, watch, type ComputedRef, type Ref, type WatchSource } from 'vue';
import type { Vec2, Vec4 } from '../util/index.type';
import { removeUnit } from '../util/index.util';
import { LazyListener, type LazyListenerOption } from '../listener/lazy.listener';

export const useLazyRef = <T>(val: Ref<T>|ComputedRef<T>, opt?: LazyListenerOption<T>) => {
    const lazy = ref(val.value);
    const lListener = new LazyListener(val.value, opt)
    lListener.onValue = v => lazy.value = v;

    watch(val, v => lListener.setValue(v))
    return readonly(lazy);
}

type CommitItem<T> = {
    get: () => T,
    source: WatchSource| WatchSource[]
} | (() => T)
type ExtractFromCommitItem<T extends CommitItem<any>> = T extends CommitItem<infer R> ? R : never
export const useCommitRefs = <T extends Record<string, CommitItem<any>>>(props: T, commit: () => void) => {
    const ctx = {} as { [K in keyof T]: Ref<ExtractFromCommitItem<T[K]>> }

    const cleanup = () => 
        Object.entries(props)
        .forEach(([k, item]) => 
            (ctx as any)[k] = ref(typeof item === 'function' ? item() : item.get())
        )


    for (const [k, item] of Object.entries(props)){
        const isFunc = typeof item === 'function';
        (ctx as any)[k] = ref(isFunc ? item() : item.get())
        if (isFunc) continue;

        watch(item.source, () => {
            (ctx[k] as Ref).value = item.get();
        })
    }

    return [ctx, {commit, cleanup}] as const
}

// type UseStateDelay<T> = Partial<{
//     duration: number;
//     delayVal: T
// }>
// export const useStateDelay = <S>(state: Ref<S>, { duration = 500, delayVal }: UseStateDelay<S> = { }) => {

//     const delayedState = ref(state.value);
//     const timeoutId = ref(NaN);
//     const lastChanged = ref(Date.now());
//     // console.log('[USE_STATE_DELAY]', DEFAULT_VAL);
//     const clear = () => {
//         if (!isNaN(timeoutId.value)) {
//             clearTimeout(timeoutId.value);
//             timeoutId.value = NaN;
//         }
//     }

//     const setDelayedState = (newState: S) => {
//         delayedState.value = newState;
//         lastChanged.value = Date.now();
//     }

//     watch(state, origin => {
//         if (delayVal === undefined || delayVal === origin) {
//             const diff = Date.now() - lastChanged.value;
//             if (diff >= duration) {
//                 setDelayedState(origin)
//                 clear();
//             } else {
//                 clear();
//                 timeoutId.value = setTimeout(() => {
//                     setDelayedState(origin)
//                     clear();
//                 }, duration - diff)
//             }
//         } else {
//             setDelayedState(origin)
//             clear();
//         }
//     })

//     onUnmounted(() => {
//         clear();
//     })
    

//     return {
//         delayedState: readonly(delayedState)
//     }
// }

export type UseResizeSizeValue = {
    outer: Vec2;
    padding: Vec4;
    rect: DOMRect;
}

export const useResize = (elRef: Ref<HTMLElement | null>) => {
    const size = ref<UseResizeSizeValue|null>(null);
    const refresh = () => {
        if (!elRef.value) {
            size.value = null;
            return
        };

        const { 
            paddingTop, paddingRight, 
            paddingBottom, paddingLeft, 
            width, height
        }  = getComputedStyle(elRef.value);
        
        const w = removeUnit(width),
            h = removeUnit(height),
            pt = removeUnit(paddingTop),
            pr = removeUnit(paddingRight),
            pb = removeUnit(paddingBottom),
            pl = removeUnit(paddingLeft);
        ;
        size.value =  {
            outer: [ w, h, ],
            padding: [ pt, pr, pb, pl ],
            rect: elRef.value.getBoundingClientRect()
        }
    }
    const observer = ref(new ResizeObserver(refresh))
    const mutation = ref(new MutationObserver(refresh))

    watch(elRef, (el, oldEl) => {
        if (oldEl) {
            observer.value.unobserve(oldEl);
            mutation.value.disconnect();
        }
        if (!el) {
            size.value = null
            return
        }

        observer.value.observe(el);
        mutation.value.observe(el, { attributes: true, })
        // update();
    })

    onUnmounted(() => {
        observer.value.disconnect();
        size.value = null
    })

    return size;
}