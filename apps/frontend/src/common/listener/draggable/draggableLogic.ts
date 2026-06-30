import type { Vec2 } from '@/common/util/index.type';
import { DragListener } from '../drag.listener';
import { DRAGGABLE_TYPE, type DraggableListener, type DraggableType } from '../draggable.listener';
import { findParent } from '@/common/util/selector.util';
import { LongPressListener } from '../longPress.listener';

export type ReleaseEventProps = Partial<{ type: DraggableType|null; }>

const {
    READY,
    DRAG_START,
    DRAG_MOVE,
    DRAG_END
    // RELEASE
} = DRAGGABLE_TYPE.constants

export abstract class DraggableLogic {
    protected draggableListener: DraggableListener;
    protected area: HTMLElement;
    protected dragListener: DragListener | null = null;
    protected selectorItem: string
    private onDownListener: ((event: MouseEvent|TouchEvent) => void) | null = null;


    constructor(area: HTMLElement, draggableListener: DraggableListener, selectorItem: string) {
        this.area = area;
        this.draggableListener = draggableListener;
        this.selectorItem = selectorItem;
    }

    // abstract setup(): void;

    setup() {
        const this_ = this;
        const { selectorItem } = this;

        this.onDownListener = event => {
            // console.log('[SETUP]', selectorItem);
            const el = (event.target as HTMLElement)
                .closest(selectorItem) as HTMLElement
            const area = findParent(this_.area, el, 2) ?? null;
            // console.log(`[${this_.area.className}]`,el, area);
            if (!area) return;

            this_.setupStartListener(event)
        }
        this.area.addEventListener('mousedown', this.onDownListener);
        this.area.addEventListener('touchstart', this.onDownListener);

    }
    protected abstract setupStartListener(event: MouseEvent | TouchEvent): void;

    protected setupDraggableListener(): boolean {
        const hold = this.draggableListener.hold;
        if (!hold) return false;
        const this_ = this;
        const draggableListener = this.draggableListener
        
        const drag = new DragListener(hold);
        this.dragListener = drag;
        drag.setup();

        drag.start = (_: Vec2, event: MouseEvent|TouchEvent) => {
            if (event instanceof MouseEvent && event.button !== 0) return;

            draggableListener.setType(DRAG_START.value())

            draggableListener.setupPointer(event, this_.selectorItem)

        }

        drag.drag = (_: Vec2, event: MouseEvent|TouchEvent) => {
            if (DRAG_MOVE.value() !== draggableListener.type)
            draggableListener.setType(DRAG_MOVE.value())
            
            draggableListener.setupPointer(event, this_.selectorItem)

            draggableListener.setupAround()
        }

        drag.end = () => this_.releaseEvent({ type: DRAG_END.value() });

        return true;
    }

    protected onDown(event: MouseEvent|TouchEvent, client?: Vec2) {
        const { draggableListener } = this;
        if (!draggableListener.hold) return

        if( event instanceof MouseEvent) {
            draggableListener.hold.dispatchEvent(new MouseEvent('mousedown', {
                clientX: client?.[0] ?? event.clientX,
                clientY: client?.[1] ?? event.clientY,
                button: event.button,
            }))
        } else if (event instanceof TouchEvent) {   
            const touchObj = new Touch({
                identifier: Date.now(),
                target: draggableListener.hold,
                clientX: client?.[0] ?? event.touches[0]!.clientX,
                clientY: client?.[1] ?? event.touches[0]!.clientY,
                radiusX: 2.5,
                radiusY: 2.5,
                rotationAngle: 10,
                force: 0.5
            })
            draggableListener.hold.dispatchEvent(new TouchEvent('touchstart', {
                cancelable: true,
                touches: [touchObj],
                targetTouches: [touchObj],
                changedTouches: [touchObj],
                view: window,
            }) )  
        }
    }

    releaseEvent({ type = null }: ReleaseEventProps = {}) {
        this.draggableListener.setType(type);
        this.draggableListener.setupAround();
        if (this.dragListener) {
            this.dragListener.cleanup();
            this.dragListener = null;
        }
    }

    cleanup() {
        if (this.onDownListener) {
            this.area.removeEventListener('mousedown', this.onDownListener)
            this.area.removeEventListener('touchstart', this.onDownListener)
            this.onDownListener = null;
        }
        if (this.dragListener) {
            this.dragListener.cleanup();
            this.dragListener = null;
        }
    }
}


export class DraggableLongPressLogic extends DraggableLogic {
    private longPressListener: LongPressListener | null = null;

    // constructor(area: HTMLElement, draggableListener: DraggableListener) {
    //     super(area, draggableListener);
    // }

    

    setupStartListener(event: MouseEvent | TouchEvent): void {
        const this_ = this;
        const { draggableListener, selectorItem } = this;
        const lp = new LongPressListener(this.area, 500);
        this_.longPressListener = lp;
        lp.setup();

        lp.start = event => {
            const el = (event.target as HTMLElement)
                .closest(selectorItem) as HTMLElement
            if (el) draggableListener.setHold(el);
            draggableListener.setType(READY.value());
        }
        lp.progress = (cur, total) => draggableListener.onProgress?.(cur / total);
        lp.cancel = () =>  this_.releaseEvent();
        lp.longPressed = event => {
            this_.setupDraggableListener()
            const preventClick = (e: Event) => {
                e.stopImmediatePropagation();
                e.preventDefault();
                e.stopPropagation();
                this_.area.removeEventListener('click', preventClick, true);
            };
            this_.area.addEventListener('click', preventClick, true);
            this_.onDown(event);
        }
        lp.onDown?.(event);
    }


    cleanup() {
        super.cleanup()
        if (this.longPressListener){
            this.longPressListener.cleanup()
            this.longPressListener = null;
        }
    }

    releaseEvent(param?: ReleaseEventProps): void {
        super.releaseEvent(param);
        if (this.longPressListener) {
            this.longPressListener.cleanup()
            this.longPressListener = null
        }
    }
}

export class DraggableDragLogic extends DraggableLogic {
    private anchorSelector: string|null = null
    private startListener: DragListener|null = null

    constructor(area: HTMLElement, draggableListener: DraggableListener, selectorItem: string, anchorSelector: string | null = null) {
        super(area, draggableListener, selectorItem);
        this.anchorSelector = anchorSelector;
    }

    protected setupStartListener(event: MouseEvent | TouchEvent): void {
        const this_ = this;
        const { draggableListener, selectorItem } = this;
        const el = (event.target as HTMLElement)
            .closest(selectorItem) as HTMLElement

        if (this.anchorSelector) {
            const anchor = el.querySelector(this.anchorSelector);
            if (anchor !== event.target) return;
        }

        draggableListener.setHold(el);

        const sListener = new DragListener(el);
        this.startListener = sListener;
        sListener.setup();

        const MAX_DIFF = 100;
        let sX = 0, sY = 0

        sListener.start = ([x, y], event) => {
            if (event instanceof MouseEvent && event.button !== 0) return;

            sX = x;
            sY = y;
            draggableListener.setType(READY.value());
        }
        
        sListener.drag = ([x, y]) => {
            if (draggableListener.type !== 'READY') return
            const diff = Math.pow(x - sX, 2) + Math.pow(y - sY, 2);
            draggableListener.onProgress?.(Math.min(1, diff / MAX_DIFF))
            
            if (diff > MAX_DIFF) {
                sListener.cleanup();
                this_.startListener = null;
                // this_.releaseEvent();
                this_.setupDraggableListener();
                // const preventClick = (e: Event) => {
                //     console.log('[STOP_EVENT]');
                //     e.stopImmediatePropagation();
                //     e.preventDefault();
                //     e.stopPropagation();
                //     this_.area.removeEventListener('click', preventClick, true);
                // }
                // console.log('[이벤트 막는 리스너 셋업]')
                // this_.area.addEventListener('click', preventClick, true);
                this_.onDown(event);
            }
        }
        sListener.end = () => {
            // console.log('[E N D_DRAGGING]')

            this_.releaseEvent()
        }
        this_.onDown(event)
    }
    cleanup(): void {
        // console.log('[CLEAN_EVENT]')
        super.cleanup()
        if (this.startListener) {
            this.startListener.cleanup();
            this.startListener = null;
        }
    }
    releaseEvent(param?: ReleaseEventProps): void {
        // console.log('[RELEASE_EVENT]')
        super.releaseEvent(param);
        if (this.startListener) {
            this.startListener.cleanup();
            this.startListener = null;
        }
    }

}