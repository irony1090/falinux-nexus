import { createConstant } from './constants.util';

export const DIRECTION = createConstant({
    HORIZONTAL: 0 as const,
    VERTICAL: 1 as const,
})
export type Direction = typeof DIRECTION._valType;

export const EDGE = createConstant({
    TOP: 0 as const,
    RIGHT: 1 as const,
    BOTTOM: 2 as const,
    LEFT: 3 as const,
})
export type Edge = typeof EDGE._valType;