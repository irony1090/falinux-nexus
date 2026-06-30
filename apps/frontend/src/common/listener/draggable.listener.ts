import { createConstant } from '../util/constants.util';
import { EDGE } from '../util/index.logic';
import type { Vec2, Vec4 } from '../util/index.type';
import { createMemoized } from '../util/index.util';
import type { DraggableLogic } from './draggable/draggableLogic';

// type DragableListeners = {
//     longPressListener: LongPressListener | null;
//     dragListener: DragListener | null;
// }
export const DRAGGABLE_TYPE = createConstant({
    READY: 'READY' as const,
    DRAG_START: 'DRAG-START' as const,
    DRAG_MOVE: 'DRAG-MOVE' as const,
    DRAG_END: 'DRAG-END' as const,
})

export type DraggableType = typeof DRAGGABLE_TYPE._valType;

const {
    DRAG_START,
    DRAG_END
    // RELEASE
} = DRAGGABLE_TYPE.constants

const {
    TOP, RIGHT, BOTTOM, LEFT
} = EDGE.constants

export const DRAGABLE_POINTER_TYPE = createConstant({
    MOUSE: 'MOUSE' as const,
    TOUCH: 'TOUCH' as const,
})
export type DragablePointerType = typeof DRAGABLE_POINTER_TYPE._valType;

type Neighbor = {
    el: HTMLElement
    distance: number;
}
// type Position = {
//     el: HTMLElement
//     rect: Vec4
//     center: Vec2
//     neighbor: Vec4<Neighbor|null>
// }
type Around = {
    target: HTMLElement
    relativeFromPointer: Vec2
    neighbor: Vec4<Neighbor|null>
}

class Position {
    el: HTMLElement
    rect: Vec4
    center: Vec2
    neighbor: Vec4<Neighbor|null>

    constructor(el: HTMLElement) {
        this.el = el;
        const rect = el.getBoundingClientRect();
        const t = rect.top,
            r = rect.left + rect.width,
            b = rect.top + rect.height,
            l = rect.left ;
        this.rect = [ t, r, b, l ];
        this.center = [(l + r) / 2, (t + b) / 2];
        this.neighbor = [null, null, null, null];
    }
}

// type NewDraggableLogics = (area: HTMLElement, this_: DraggableListener) => DraggableLogic|Array<DraggableLogic>
export class DraggableListener {
    private areas: Map<HTMLElement, Array<DraggableLogic>> = new Map();
    private rects = new Map<HTMLElement, Vec4>();
    private itemPositions = new Map<HTMLElement, Position>();
    private pointer: Vec2|null = null;
    // private newStartDraggable: NewDraggableLogic
    // private around: Around|null = null;
    private area: HTMLElement|null = null;
    public type: DraggableType|null = null;
    public hold: HTMLElement|null = null;
    public dragItemSelector: string|null = null;


    public onHold: ((el: HTMLElement|null) => void) | undefined;
    public onHoldIdx: ((idx: number) => void) | undefined;
    public onType: ((type: DraggableType|null) => void) | undefined;
    public onProgress: ((progress: number) => void) | undefined;
    public onPointer: ((pointer: Vec2|null, areaRect: Vec4|undefined) => void) | undefined;
    public onArea: ((el: HTMLElement|null) => void) | undefined;
    public onAround: ((around: Around|null) => void) | undefined;
    
    // constructor(newStartDraggable: NewDraggableLogic) {
    //     this.newStartDraggable = newStartDraggable
    // }

    setTypeNull() {
        this.setType(null);
    }
    
    setType(type: DraggableType|null) {
        const diff = type !== this.type;
        this.type = type;
        if (diff) this.onType?.(type);
        switch(type) {
            case DRAG_START.value():
                this.initRects();
                break;
            case DRAG_END.value():
                this.onProgress?.(0);
                this.onHoldIdx?.(this.getHoldIdx());
                this.cleanupRects();
                break;
            case null:
                this.onProgress?.(0);
                this.setPointer(null);
                this.setHold(null);
                this.onHoldIdx?.(NaN);
                this.cleanupRects();
                this.dragItemSelector = null;
                break;
        }
    }

    getHoldIdx(): number {
        if (!this.hold || this.itemPositions.size === 0) return NaN;
        let idx = 0;
        for (let [el] of this.itemPositions) {
            if (this.hold === el) break;
            else idx++;
        }
        if (idx >= this.itemPositions.size)
            idx = NaN;
            
        return idx;
    }

    setHold(hold: HTMLElement|null) {
        const diff = hold !== this.hold;
        this.hold = hold;

        if (diff) this.onHold?.(hold)
    }
    setPointer(pointer: Vec2|null) {
        this.pointer = pointer;

        const t = this.contains();
        if (t !== this.area) this.setArea(t);

        this.onPointer?.(pointer, t ? this.rects.get(t) : undefined);
    }

    addArea(el: HTMLElement, logic: DraggableLogic|Array<DraggableLogic>) {
        if (!this.areas.has(el)) {
            const logics = Array.isArray(logic) ? logic : [logic];
            logics.forEach(l => l.setup());
            this.areas.set(el, logics);
            // const areaLogic = this.newStartDraggable(el, this)
            // const areaLogic = new DragableLongPressLogic(el, this);
            // areaLogic.setup();
            // this.areas.set(el, areaLogic);
        }
    }

    removeArea(el: HTMLElement) {
        const l = this.areas.get(el);
        if (l) {
            l.forEach(l => l.cleanup());
            // l.cleanup();
            this.areas.delete(el);
        }
    }

    setupAround() {
        if (!this.pointer || this.itemPositions.size === 0){
            this.setAround(null);
            return undefined;
        }

        const [px, py] = this.pointer;

        let minDistance = Infinity;
        let closest: Position | null = null
        for (const [_, pos] of this.itemPositions) {
            const [tx, ty] = pos.center
            const distance = Math.pow(px - tx, 2) + Math.pow(py - ty, 2)
            if (distance < minDistance) {
                minDistance = distance;
                closest = pos;
            }
        }

        if (closest)
            this.setAround({
                target: closest.el,
                relativeFromPointer: [px - closest.center[0], py - closest.center[1]],
                neighbor: closest.neighbor,
          });

    }

    setupPointer(event: MouseEvent|TouchEvent, selectorItem?: string){
        let vec2: Vec2;
        if (event instanceof MouseEvent)
            vec2 = [event.clientX, event.clientY];
            // return [event.clientX, event.clientY];
        else {
            const touch = event.touches[0]!;
            vec2 = [touch.clientX, touch.clientY];
            // return [touch.clientX, touch.clientY];
        }

        this.dragItemSelector = selectorItem ? selectorItem : null;

        this.setPointer(vec2);
    }

    private setAround(around: Around|null) {
        // this.around = around;
        this.onAround?.(around);
    }
    private setArea(area: HTMLElement|null) {
        const diff = area !== this.area;
        this.area = area;
        if (diff) {
            this.onArea?.(area);
            this.initItemRects();
        }

    }
    
    private initItemRects() {
        this.cleanupItemPositions();
        if (!this.area) return;

        const memo = createMemoized((el: HTMLElement) => new Position(el));
        const items = this.dragItemSelector ? this.area.querySelectorAll(this.dragItemSelector) : [];

      // setupDirection을 밖으로 추출
        const setupDirection = (
            current: Position,
            next: Position,
            nextEl: HTMLElement,
            isHorizontal: boolean,
            distance: number
        ) => {
            if (distance === 0) return;

            const dirIndex = isHorizontal
                ? (distance > 0 ? RIGHT.value() : LEFT.value())
                : (distance > 0 ? BOTTOM.value() : TOP.value());

            const orgDistance = current.neighbor[dirIndex]?.distance ?? NaN;
            const absDistance = Math.abs(distance);

            if (isNaN(orgDistance) || absDistance < orgDistance) {
                current.neighbor[dirIndex] = { el: nextEl, distance: absDistance };

                const oppositeDirIndex = isHorizontal
                    ? (dirIndex === RIGHT.value() ? LEFT.value() : RIGHT.value())
                    : (dirIndex === TOP.value() ? BOTTOM.value() : TOP.value());

                next.neighbor[oppositeDirIndex] = { el: current.el, distance: absDistance };
            }
        };

        // 한 번의 순회로 처리
        items.forEach((el_, i) => {
            const el = el_ as HTMLElement;
            const current = memo.get(el);

            const nextEl = (i < items.length - 1) ? items[i + 1] as HTMLElement : null;
            if (nextEl) {
                const next = memo.get(nextEl);
                const distanceX = next.center[0] - current.center[0];
                const distanceY = next.center[1] - current.center[1];

                setupDirection(current, next, nextEl, true, distanceX);
                setupDirection(current, next, nextEl, false, distanceY);
            }

            // 바로 저장
            this.itemPositions.set(el, current);
        });
    }
    private cleanupItemPositions() {
        this.itemPositions.clear();
    }

    private initRects() {
        this.cleanupRects();
        this.areas.forEach((_, k) => {
            const rect = k.getBoundingClientRect();
            this.rects.set(k, [
                rect.top, 
                rect.left + rect.width, 
                rect.top + rect.height, 
                rect.left
            ]);
        })
    }
    private cleanupRects() {
        this.cleanupItemPositions();
        this.rects.clear();
    }

    

    private contains(): HTMLElement|null {
        if (!this.pointer) return null;
        
        for (const [el, rect] of this.rects) {
            const [top, right, bottom, left] = rect;
            const [px, py] = this.pointer;
            if (px >= left && px <= right && py >= top && py <= bottom)
                return el;

        }
        return null;
    }
}
