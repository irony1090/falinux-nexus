import type { Ref } from 'vue'

export type QueryOptions = Partial<{
    enabled: Ref<boolean>
    refetchInterval: Ref<number>
}>

export type ToRefs<T> = {
    [K in keyof T]: Ref<T[K]>
}