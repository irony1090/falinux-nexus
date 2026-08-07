import { GroupedSet } from '@/common/util/groupedSet.util';
import { LifecycleRegistry } from '@/common/util/lifecycle/lifecycle.util';
import { inject, onBeforeUnmount, provide, ref, watch, type Ref } from 'vue';

const STORE_KEY = Symbol('RESIZE_GROUP_STORE_KEY');

type EventCallback = ((el: HTMLElement) => void)

export const provideResizeGroupStore = () => {

    const flag = ref(false);

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

    const callbacks = new GroupedSet<HTMLElement, () => void>()

    const predict: EventCallback = el => 
        callbacks.forEach(el, v => v())

    const resizeObserver = new ResizeObserver(entities => {
        entities.forEach(entry => predict(entry.target as HTMLElement) );
    });
    const mutationObserver = new MutationObserver((record) => {
        record.forEach(rcd => predict(rcd.target as HTMLElement) );
    });

    const cycle = new LifecycleRegistry<HTMLElement>({
        init: el => {
            resizeObserver.observe(el);
            mutationManager.observe(el);
        },
        release: el =>{
            resizeObserver.unobserve(el);
            mutationManager.unobserve(el);
        },
    })
    cycle.on('registered', _ => {
        if (flag.value) cycle.initAll();
    })
    
    const registryElement = (el: HTMLElement, cb: () => void) => {
        callbacks.put(el, cb);
        cycle.put(el);
    }
    const releaseElement = (el: HTMLElement, cb: () => void) => {
        callbacks.remove(el, cb);
        if (!callbacks.has(el)) cycle.remove(el)
    }

    const ctx = {
        registryElement,
        releaseElement,
        flag,
    };

    watch(flag, val => {
        if (val) cycle.initAll();
        else cycle.releaseAll()
    }, { immediate: true })


    provide(STORE_KEY, ctx);
    return ctx;
}

type Context = ReturnType<typeof provideResizeGroupStore>;
const useResizeGroupStore = () => {
    const context = inject<Context>(STORE_KEY)
    if (!context) throw new Error('useResizeGroupStore is not provided');
    return context;
}

export const useResizeCallback = (
    elRef: Ref<HTMLElement|undefined|null>, 
    resizeCallBack: () => void,
) => {
    const { registryElement, releaseElement } = useResizeGroupStore();

    // const elRef = ref<HTMLElement>();
    
    watch(elRef, (val, oldVal) => {
        if (val) {
            registryElement(val, resizeCallBack)
        } else if (oldVal) {
            releaseElement(oldVal, resizeCallBack)
        }
    })

    onBeforeUnmount(() => {
        if (elRef.value) releaseElement(elRef.value, resizeCallBack)
    })


    const context = { elRef }

    return context

}