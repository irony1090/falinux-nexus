import { useResize, type UseResizeSizeValue } from '@/common/hook/vue.hook';
import { computed, inject, provide, ref, type Ref } from 'vue';

const APP_HEAD_STORE_KEY = 'AppHeadStore';
type ContextType = {
    vElRef: Ref<any>;
    size: Ref<UseResizeSizeValue | null>;
    backUrl: Ref<string|undefined>;
}

export const provideAppHead = () => {
    const vElRef = ref<any>(null);
    const elRef = computed<HTMLElement | null>(() => vElRef.value?.$el ?? null);
    const size = useResize(elRef);
    const backUrl = ref<string>();
    const context: ContextType = { size, vElRef, backUrl };
    provide(APP_HEAD_STORE_KEY, context);
    return context;
}

// type ContextType = ReturnType<typeof provideAppHead>

export const useAppHead = () => {
    const context = inject<ContextType>(APP_HEAD_STORE_KEY)!;
    if (!context) throw new Error('AppHeadStore is not provided');
    return context;
}