import { onMounted, onUnmounted, readonly, ref, watch, type ComputedRef, type Ref, type WatchSource } from 'vue';
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

export type UseResizeSizeValue = {
    outer: Vec2;
    padding: Vec4;
    rect: DOMRect;
}

type UseResizeOptions = {
    // ResizeObserver/MutationObserver(attributes)는 대상의 "크기"만 감지하고 "위치"(rect) 변화는
    // 못 잡는다 — 예: max-width에 걸려 폭은 고정된 채 중앙정렬 offset만 바뀌는 경우.
    // rect(위치)까지 정확해야 하는 호출부에서만 켤 것 (모든 호출부에 기본으로 켜면 리스너 부담이 커짐).
    watchWindowResize?: boolean;
}

export const computeResizeSize = (el: HTMLElement|null):UseResizeSizeValue|null => {
    if (!el) {
        return null
    };

    const { 
        paddingTop, paddingRight, 
        paddingBottom, paddingLeft, 
        width, height
    }  = getComputedStyle(el);
    
    const w = removeUnit(width),
        h = removeUnit(height),
        pt = removeUnit(paddingTop),
        pr = removeUnit(paddingRight),
        pb = removeUnit(paddingBottom),
        pl = removeUnit(paddingLeft);
    
    return {
        outer: [ w, h, ],
        padding: [ pt, pr, pb, pl ],
        rect: el.getBoundingClientRect()
    }
}

export const useResize = (elRef: Ref<HTMLElement | null>, { watchWindowResize = false }: UseResizeOptions = {}) => {
    const size = ref<UseResizeSizeValue|null>(null);
    const refresh = () => {
        size.value = computeResizeSize(elRef.value)
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

    if (watchWindowResize) {
        onMounted(() => window.addEventListener('resize', refresh));
        onUnmounted(() => window.removeEventListener('resize', refresh));
    }

    onUnmounted(() => {
        observer.value.disconnect();
        size.value = null
    })

    return size;
}

// 여러 엘리먼트의 scroll/resize/속성변화를 한 콜백으로 감지. targets가 바뀌면 필요한 것만 추가/해제.
export const useElementsChange = (
    targets: Ref<(HTMLElement | null | undefined)[]> | ComputedRef<(HTMLElement | null | undefined)[]>,
    onChange: () => void
) => {
    const resizeObserver = new ResizeObserver(onChange);
    const mutationObserver = new MutationObserver(onChange);
    const attached = new Set<HTMLElement>();

    const sync = () => {
        const current = new Set(targets.value.filter((el): el is HTMLElement => !!el));

        for (const el of attached) {
            if (current.has(el)) continue;
            el.removeEventListener('scroll', onChange);
            resizeObserver.unobserve(el);
            attached.delete(el);
        }
        for (const el of current) {
            if (attached.has(el)) continue;
            el.addEventListener('scroll', onChange, { passive: true });
            resizeObserver.observe(el);
            mutationObserver.observe(el, { attributes: true });
            attached.add(el);
        }
    }

    watch(targets, sync, { immediate: true });

    onUnmounted(() => {
        resizeObserver.disconnect();
        mutationObserver.disconnect();
        for (const el of attached) el.removeEventListener('scroll', onChange);
        attached.clear();
    })
}
