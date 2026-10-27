// 찰떡 입력칸 공용 속성 — CtTextField·CtSelect·CtAutocomplete·CtTextarea (모양 = styles/ct-field.scss, REF-ct-field.md)
export const ctFieldProps = {
    size: {             // sm = 32px(라벨은 칸 밖) / md = 40px(라벨이 선에 걸침)
        type: String as () => 'sm' | 'md',
        default: 'md',
    },
    wall: {             // 곧게 할 끝 — 부모 벽에 닿거나 옆 칸과 붙는 쪽
        type: String as () => 'l' | 'r' | 'lr',
        default: undefined,
    },
}

export const ctFieldClass = (p: { size: string; wall?: string }) => ['ct-field', `ct-field--${p.size}`, p.wall && `ct-field--wall-${p.wall}`];
