import { settings } from '@/feature/settings/store/settings.store';
import { computed } from 'vue';

export type TileBarAt = 'top' | 'bottom';

// 타일 메뉴 줄 위치 = 설정값(브라우저별) — REF-tile-frame.md · REF-settings-footer.md
export const tileBarAt = computed(() => settings.value.tileBarAt);
