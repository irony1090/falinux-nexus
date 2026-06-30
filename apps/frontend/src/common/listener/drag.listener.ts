import type { Vec2 } from '../util/index.type'

// type DragType = 'START' | 'DRAG' | 'END'
type DragEvent = (offset: Vec2, event: MouseEvent | TouchEvent) => void

export class DragListener {
    private element: HTMLElement
    start: DragEvent|undefined
    drag: DragEvent|undefined
    end: DragEvent|undefined
    rect: DOMRect|undefined

    private onDown: ((event: MouseEvent|TouchEvent) => void)|undefined
    private onMove: ((event: MouseEvent|TouchEvent) => void)|undefined
    private onUp: ((event: MouseEvent|TouchEvent) => void)|undefined

    constructor(element: HTMLElement) {
        this.element = element
    }

    setup(): DragListener {
        this.cleanup();

        this.onMove = event => {
            if (!this.rect) return;
            let x, y;
            const rect = this.rect;
            // const rect = this.element.getBoundingClientRect();
            if (event instanceof MouseEvent) {
                x = event.clientX - rect.left;
                y = event.clientY - rect.top;
            } else {
                x = event.touches[0]!.clientX - rect.left;
                y = event.touches[0]!.clientY - rect.top;
            }
            this.drag?.([x,y], event)
        }

        this.onUp = event => {
            let x, y;
            const rect = this.rect ?? this.element.getBoundingClientRect();
            if (event instanceof MouseEvent) {
                x = event.clientX - rect.left;
                y = event.clientY - rect.top;
                window.removeEventListener('mousemove', this.onMove as any)
                window.removeEventListener('mouseup', this.onUp as any)
            } else {
                x = event.changedTouches[0]!.clientX - rect.left;
                y = event.changedTouches[0]!.clientY - rect.top;
                window.removeEventListener('touchmove', this.onMove as any)
                window.removeEventListener('touchend', this.onUp as any)
            }
            // console.log('[DRAG_LISTENER_UP]')
            this.end?.([x,y], event)
            this.rect = undefined
        }

        this.onDown = event => {
            // console.log('[DRAG_LISTENER_DOWN]')
            event.preventDefault();
            let x, y;
            this.rect = this.element.getBoundingClientRect();
            if (event instanceof MouseEvent) {
                x = event.clientX - this.rect.left;
                y = event.clientY - this.rect.top;
                window.addEventListener('mousemove', this.onMove as any);
                window.addEventListener('mouseup', this.onUp as any);
            } else {
                // const rect = this.element.getBoundingClientRect();
                x = event.touches[0]!.clientX - this.rect.left;
                y = event.touches[0]!.clientY - this.rect.top;
                window.addEventListener('touchmove', this.onMove as any);
                window.addEventListener('touchend', this.onUp as any);
            }
            this.start?.([x,y], event);
        }

        this.element.addEventListener('mousedown', this.onDown);
        this.element.addEventListener('touchstart', this.onDown);

        return this
    }

    cleanup() {
        if (this.onDown) {
            this.element.removeEventListener('mousedown', this.onDown);
            this.element.removeEventListener('touchstart', this.onDown);
            this.onDown = undefined;
        }

        if (this.onMove) {
            window.removeEventListener('mousemove', this.onMove);
            window.removeEventListener('touchmove', this.onMove);
            this.onMove = undefined;
        }

        if (this.onUp) {
            window.removeEventListener('mouseup', this.onUp);
            window.removeEventListener('touchend', this.onUp);
            this.onUp = undefined;
        }
    }
}