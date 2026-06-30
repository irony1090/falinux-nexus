import { useResize } from '@/common/hook/vue.hook';
import { computed, inject, provide, ref } from 'vue';

const APP_NAV_STORE_KEY = 'AppNavStore';


export const provideAppNav = () => {
    const open = ref(false);
    // const vElRef = ref<any>(null);
    const elRef = ref<any>(null);
    // const elRef = computed<HTMLElement | null>(() => vElRef.value?.$el ?? null);
    const _size = useResize(elRef);

    const size = computed(() => {
        if (open.value) return _size.value;
        else return null;
    })
    const context = { size, elRef, open };
    provide(APP_NAV_STORE_KEY, context);
    return context;
}

type ContextType = ReturnType<typeof provideAppNav>

export const useAppNav = () => {
    const context = inject<ContextType>(APP_NAV_STORE_KEY)!;
    if (!context) throw new Error('AppFooterStore is not provided');
    return context;
}