import type { ComponentPublicInstance, Ref } from 'vue';

// 템플릿 함수 ref -> 루트 요소. 요소 하나를 감싼 컴포넌트(TileShell·TileScroll)에 ref를 걸 때
export const elRef = (r: Ref<HTMLElement | undefined>) => (c: Element | ComponentPublicInstance | null) => {
    r.value = ((c && '$el' in c ? c.$el : c) ?? undefined) as HTMLElement | undefined;
};
