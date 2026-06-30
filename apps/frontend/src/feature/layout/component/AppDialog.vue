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
        <v-divider />
        <v-card-actions class="actions">
            <v-btn v-for="{ view, action, color, variant }, i in buttons" 
                :key="i"
                :color="color"
                :variant="variant"
                @click="action"
            >
                {{ view }}
            </v-btn>
        </v-card-actions>
    </v-card>
</v-dialog>
</template>

<script lang="ts" setup>
import { useAppDialog, type DialogButton } from '@/feature/layout/store/appDialog.store';
import { computed, watch } from 'vue';
import { VDialog, VCard, VCardTitle, VCardText, VCardActions, VDivider, VBtn } from 'vuetify/components'
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


const buttons = computed(() => {
    return (dialog.value?.buttons ?? []).map((el) => {
        const newEl = {...el} as DialogButton & {
            variant: VBtn['$props']['variant'],
            color: VBtn['$props']['color'],
        };
        if (!el.type) {
            newEl.type = dialog.value?.type
        }

        switch(newEl.type) {
            case 'info':
                newEl.variant = 'text';
                newEl.color = undefined;
                break;
            case 'warning':
                newEl.variant = 'outlined';
                newEl.color = 'warning';
                break;
            case 'error':
                newEl.variant = 'flat';
                newEl.color = 'error';
                break;
            default:
                newEl.variant = 'plain';
                newEl.color = 'error';
        }

        return newEl;
    })
})


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
.actions {
    display: flex;
    align-items: center;
    justify-content: end;
}
</style>