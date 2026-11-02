<script setup lang="ts">
import { computed } from 'vue';
import MbTileFrame, { type MbCrumb } from '@/feature/member/component/MbTileFrame.vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { Placed } from '@/feature/member/util/memberGrid.util';
import { buildTileKey, commonTileKey, devTileKey, updTargetOf } from '../util/updGrid.util';
import UpdBuildTile from './UpdBuildTile.vue';
import UpdCommonTile from './UpdCommonTile.vue';
import UpdDeviceTile from './UpdDeviceTile.vue';
import UpdFilesTile from './UpdFilesTile.vue';
import UpdStepsTile from './UpdStepsTile.vue';
import UpdTextTile from './UpdTextTile.vue';

// updater 페이지 상세 타일 — 경로·끌기 표시·몸통. 틀 = MbTileFrame(회원 페이지와 같음) — REF-update-deploy.md "화면"
const props = defineProps<{ placed: Placed }>();
const { tiles, open, activate, openParent, pickTab } = useMemberBoard();

const tile = computed(() => props.placed.tile);
const t = computed(() => updTargetOf(tile.value.key));

// 경로의 앞 조각 = 그 타일로. 닫혀 있으면 연 곳(부모)이면 자식 바로 앞에, 아니면 새로 연다
const goTo = (key: string) => {
    if (tiles.value.some(x => x.key === key)) return activate(key);
    if (key === tile.value.parent) return openParent(tile.value);
    open(key);
};

const META = {
    device: { icon: 'mdi-folder-network', color: 'info', label: '장비 타일' },
    build: { icon: 'mdi-package-variant', color: 'primary', label: '빌드 타일' },
    files: { icon: 'mdi-file-multiple-outline', color: 'primary', label: 'Files 타일' },
    steps: { icon: 'mdi-format-list-numbered', color: 'primary', label: 'steps.json 타일' },
    common: { icon: 'mdi-package-variant-closed', color: 'share', label: '공용 묶음 타일' },
    text: { icon: 'mdi-file-document-outline', color: 'primary', label: 'Script 타일' },
} as const;

const crumbs = computed<MbCrumb[]>(() => {
    const { kind, mainKey, ref, path } = t.value;
    if (kind === 'device') return [{ label: '장비', go: () => pickTab('devices') }, { label: mainKey }];
    if (kind === 'common') return [{ label: '공용 묶음', go: () => pickTab('commons') }, { label: ref!.owner }, { label: ref!.version }];
    const owner = ref!.kind === 'build'
        ? { label: ref!.owner, go: () => goTo(devTileKey(ref!.owner)) }
        : { label: ref!.owner, go: () => pickTab('commons') };
    const bundleKey = ref!.kind === 'build' ? buildTileKey(ref!.owner, ref!.version) : commonTileKey(ref!.owner, ref!.version);
    if (kind === 'build') return [{ label: '장비', go: () => pickTab('devices') }, owner, { label: ref!.version }];
    const last = kind === 'files' ? 'Files' : kind === 'steps' ? 'steps.json' : path!;
    return [owner, { label: ref!.version, go: () => goTo(bundleKey) }, { label: last }];
});
const name = computed(() => crumbs.value.at(-1)?.label ?? '');
const chip = computed(() => ({ icon: META[t.value.kind].icon, color: META[t.value.kind].color, name: name.value }));
</script>

<template>
<mb-tile-frame class="UpdTileView" :placed="placed" :crumbs="crumbs" :chip="chip" :label="META[t.kind].label">
    <upd-device-tile v-if="t.kind === 'device'" :tile="tile" :main-key="t.mainKey" />
    <upd-build-tile v-else-if="t.kind === 'build'" :tile="tile" :bundle="t.ref!" />
    <upd-files-tile v-else-if="t.kind === 'files'" :tile="tile" :bundle="t.ref!" />
    <upd-steps-tile v-else-if="t.kind === 'steps'" :tile="tile" :bundle="t.ref!" />
    <upd-common-tile v-else-if="t.kind === 'common'" :tile="tile" :bundle="t.ref!" />
    <upd-text-tile v-else :tile="tile" :bundle="t.ref!" :path="t.path!" />
</mb-tile-frame>
</template>
