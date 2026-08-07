import { type UseResizeSizeValue } from '@/common/hook/vue.hook';
import { inject, provide, ref, type Ref } from 'vue';
import type { VAppBar } from 'vuetify/components';

const APP_HEAD_STORE_KEY = 'AppHeadStore';
type ContextType = {
    vElRef: Ref<any>;
    size: Ref<UseResizeSizeValue | null>;
    backUrl: Ref<string|undefined>;
}

export const provideAppHead = () => {
    const vElRef = ref<InstanceType<typeof VAppBar>>();

    const size = ref<UseResizeSizeValue|null>(null)
    // const size = useResize(elRef);
    const backUrl = ref<string>();
    const context: ContextType = { size, vElRef, backUrl };
    provide(APP_HEAD_STORE_KEY, context);
    return context;
}


export const useAppHead = () => {
    const context = inject<ContextType>(APP_HEAD_STORE_KEY)!;
    if (!context) throw new Error('AppHeadStore is not provided');
    return context;
}