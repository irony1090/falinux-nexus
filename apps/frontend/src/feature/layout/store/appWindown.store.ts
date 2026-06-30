import { type UseResizeSizeValue } from '@/common/hook/vue.hook';
import { getSizeConfig, SIZE_RECORD, type SizeConfig } from '@/common/layout/size.layout';
import type { Vec2, Vec4 } from '@/common/util/index.type';
import { removeUnit } from '@/common/util/index.util';
import { inject, onBeforeUnmount, onMounted, provide, ref, watch, type Ref } from 'vue';

const APP_WINDOW_STORE_KEY = 'AppWindowStore';
type ContextType = {
    size: Ref<UseResizeSizeValue & {inner: Vec2} | null>;
    sizeConfig: Ref<SizeConfig>;
    client: Ref<HTMLElement|null>
    // baseFontSize: ComputedRef<number>;
}

export const provideAppWindow = () => {
    const size = ref<UseResizeSizeValue& {inner: Vec2}|null>(null);
    const sizeConfig = ref<SizeConfig>(SIZE_RECORD.MD);
    const client = ref<HTMLElement|null>(null);
    // const baseFontSize = computed(() => sizeConfig.value?.baseFontSize ?? SIZE_RECORD.SM.baseFontSize )
    const context: ContextType = { size, sizeConfig, client };
    const muObserver = ref<MutationObserver>();

    const onResize = () => {
        // const {innerWidth, innerHeight} = window;
        const {clientWidth, clientHeight} = document.documentElement;
        const { width = 0, height = 0 } = window.visualViewport ?? {};
        
        const s = client.value ? getComputedStyle(client.value) : undefined;
        const padding: Vec4 = s 
            ? [
                removeUnit(s.getPropertyValue('--v-layout-top').trim(), 0),
                removeUnit(s.getPropertyValue('--v-layout-right').trim(), 0),
                removeUnit(s.getPropertyValue('--v-layout-bottom').trim(), 0),
                removeUnit(s.getPropertyValue('--v-layout-left').trim(), 0)
            ]
            : [0, 0, 0, 0]
        // console.log('Window resized', {outerWidth, outerHeight, innerWidth, innerHeight});
        sizeConfig.value = getSizeConfig(innerWidth);
        size.value = {
            inner: [width, height],
            outer: [ clientWidth, clientHeight ],
            rect: document.documentElement.getBoundingClientRect(),
            padding
        }
    }

    watch(client, (val) => {
        if (val) {
            muObserver.value = new MutationObserver( onResize )
            muObserver.value.observe(val, { attributes: true, attributeFilter: ['style']})
        } else {
            muObserver.value?.disconnect();
            muObserver.value = undefined;
        }
    })

    watch(sizeConfig, val => {
        if (!val) return;
        document.querySelector('html')?.setAttribute('data-size-grade', val.grade);
    }, {immediate: true})

    onMounted(() => {
        onResize();
        window.addEventListener('resize', onResize);
    })
    onBeforeUnmount(() => {
        window.removeEventListener('resize', onResize);
        size.value = null;
    })
    
    provide(APP_WINDOW_STORE_KEY, context);
    return context;
}

export const useAppWindow = () => {
    const context = inject<ContextType>(APP_WINDOW_STORE_KEY)!;
    if (!context) throw new Error('AppWindowStore is not provided');
    return context;
}