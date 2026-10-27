<template>
<v-dialog 
    max-width="396"
    min-width="296"
    :persistent="persistent"
    v-model="open"
    
>
    <v-card
        :loading="delayLoading"
        :disabled="delayLoading"
        class="ct-card"
    >
        <v-card-title class="title">
            <v-icon 
                :color="iconAttrs.color" 
                :size="iconSize"
            >{{ iconAttrs.view }}</v-icon>
            <h5 v-if="dialog?.title">{{ dialog.title }}</h5>
        </v-card-title>
        <v-divider />
        <v-card-text>
            <p v-for="content,i in contents" :key="i" v-html="content" />
        </v-card-text>
        <div class="ct-actions">
            <span class="ct-grp ct-grp--r">
                <button v-for="{ view, action, cls }, i in buttons" :key="i" type="button" class="ct-seg" :class="cls" @click="action">{{ view }}</button>
            </span>
        </div>
    </v-card>
</v-dialog>
</template>

<script lang="ts" setup>
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { computed, watch } from 'vue';
import { VDialog, VCard, VCardTitle, VCardText, VDivider } from 'vuetify/components'
import { useLazyRef } from '@/common/hook/vue.hook';
import { useAppWindow } from '@/feature/layout/store/appWindown.store';
// import { useStateDelay } from '@/common/hook/vue.hook';

const { sizeConfig } = useAppWindow();
const iconSize = computed(() => sizeConfig.value.baseFontSize * 1.5 );

const { 
    dialog, 
    open,
    loading
} = useAppDialog();

const delayLoading = useLazyRef(loading, { startLazy: v => !v, time: 750 })
// const { delayedState: delayLoading } = useStateDelay(loading, { delayVal: false, duration: 750 })

const contents = computed<Array<string>>(() => {
    const content = dialog.value?.content;
    if (!content) return []

    return Array.isArray(content) ? content : [content]
})


// 버튼 종류 = 찰떡 칸 색(REF-design-chaltteok.md). 종류가 없으면 창 종류를 따른다
const SEG_CLS: Record<string, string> = { info: '', warning: 'ct-seg--warning', error: 'ct-seg--danger' };
const buttons = computed(() => (dialog.value?.buttons ?? []).map((el) => {
    const type = el.type ?? dialog.value?.type;
    return { ...el, type, cls: type ? SEG_CLS[type] : 'ct-seg--error' };
}))


const iconAttrs = computed(() => {
    switch(dialog.value?.type) {
        case 'info':
            return {
                view: 'mdi-lightbulb-on-outline',
                color: undefined
            };
        case 'warning':
            return {
                view: 'mdi-alert-circle-outline',
                color: 'warning',
            };
        case 'error':
            return {
                view: 'mdi-alert-outline',
                color: 'error',
            };
        default:
            return {
                view: 'mdi-alert-decagram-outline',
                color: 'error'
            };
    }
})

const onESC = (event: KeyboardEvent) => {
    if (loading.value) {
        return ;
    }
    if (['Escape', 'Enter', ' '].includes(event.key)) {
        event.preventDefault();
        const k = event.key === ' ' ? 'Whitespace' : event.key;
        const arr = dialog.value?.buttons.filter(({bindButton}) => {
            return (Array.isArray(bindButton) ? bindButton : [bindButton])
            .some(btnCode => btnCode === k)
        })
        if ((arr ?? []).length > 0) {
            arr!.forEach(({action}) => action());
            return;
        }
        open.value = false;
    }
}

const persistent = computed(() => {
    return dialog.value?.persistent
})

watch(open, val => {
    if (val) {
        window.addEventListener('keydown', onESC);
    } else {
        window.removeEventListener('keydown', onESC);
    }
}, { immediate: true })

</script>

<style lang="scss" scoped>

.title {
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
}
</style>