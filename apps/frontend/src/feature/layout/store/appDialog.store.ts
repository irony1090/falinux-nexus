import { computed, inject, provide, readonly, ref, watch } from 'vue';

const APP_DIALOG_STORE_KEY = 'AppDialogStore';

type DialogType = 'info' | 'warning' | 'error'
type EventButton = 'Escape' | 'Enter' | 'Whitespace';
export type DialogButton = {
    type?: DialogType
    bindButton?: EventButton | Array<EventButton>
    view: string;
    action: () => void;
}
type AppDialogProps = {
    type: DialogType
    persistent: boolean;
    title: string | null;
    content: string|Array<string>;
    buttons: Array<DialogButton>
}

type SetDialogProps = 
    (
        Omit<Partial<AppDialogProps>, 'content'> 
        & Pick<AppDialogProps, 'content'>
    ) 
    | AppDialogProps['content']

export const provideAppDialog = () => {
    const dialog_ = ref<AppDialogProps|null>(null);
    const dialog = readonly(dialog_);

    const loading = ref(false);

    const promise = ref<Record<'resolve'|'reject', () => void>|null>(null);

    const open = computed({
        get: () => !!dialog.value,
        set: val => {
            if (!val) {
                dialog_.value = null;
            }
        }
    })

    const openDialog = (props: SetDialogProps): Promise<void> => {
        const { 
            content, 
            type = 'info',
            title = null,
            persistent = true,
            buttons = [{
                view: '닫기',
                action: () => open.value = false,
            }]

        } = typeof props === 'string' || Array.isArray(props) ? { content: props, } : props

        dialog_.value = {
            type, title,
            content, buttons,
            persistent
        };
        
        return new Promise<void>((resolve, reject) => {
            promise.value = { resolve, reject };
        }).finally(() => promise.value = null)
    }

    watch(open, val => {
        if (!val) {
            promise.value?.resolve()
            if (loading.value) {
                loading.value = false;
            }
        }
    })
    const ctx = {
        open,
        dialog,
        loading,
        openDialog,
    } 

    provide(APP_DIALOG_STORE_KEY, ctx);

    return ctx;
}

export const useAppDialog = () => {
    const context = inject<ReturnType<typeof provideAppDialog> >(APP_DIALOG_STORE_KEY)!;
    if (!context) throw new Error('AppDialogStore is not provided');
    return context
}