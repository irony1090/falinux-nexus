<script lang="ts" setup>
// 끊김 표시: 로그인 중 소켓이 SHOW_MS 넘게 연결 안 됨이면 위 가운데 알약 — 잠깐 끊겼다 붙는 건 안 보인다(REF-socket-reconnect.md)
import { onBeforeUnmount, ref, watch } from 'vue'
import type { Connect } from '@/common/util/index.type'

const props = defineProps<{ status: Connect['Status']; on: boolean }>()

const SHOW_MS = 1500
const shown = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

watch(() => props.on && props.status !== 'CONNECTED', lost => {
    clearTimeout(timer)
    if (!lost) shown.value = false
    else timer = setTimeout(() => shown.value = true, SHOW_MS)
}, { immediate: true })

onBeforeUnmount(() => clearTimeout(timer))
</script>

<template>
<div v-if="shown" class="SocketLostPill" role="status">
    <span class="SocketLostPill__dot" />연결 끊김 · 다시 연결 중
</div>
</template>

<style lang="scss" scoped>
.SocketLostPill {
    position: fixed;
    top: 8px;
    left: 50%;
    translate: -50% 0;
    z-index: 3000;
    pointer-events: none;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    border-radius: 999px;
    background: rgb(var(--v-theme-warning));
    color: rgb(var(--v-theme-on-warning));
    @include vars.font('label-medium');
}
.SocketLostPill__dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
    animation: socket-lost-blink 1s ease-in-out infinite alternate;
}
@keyframes socket-lost-blink { to { opacity: .25; } }
</style>
