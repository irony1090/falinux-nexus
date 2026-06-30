
// type Options = Partial<{

import { TimeElapse } from '../util/index.util'

// }>
export class LongPressListener {
    private element: HTMLElement
    private startAnimation = false;
    private delay: number
    private elapse: TimeElapse = new TimeElapse()

    public start: ((event: MouseEvent|TouchEvent) => void)|undefined
    public longPressed: ((event: MouseEvent|TouchEvent) => void)|undefined
    public progress: ((current: number, total: number) => void)|undefined
    public cancel: ((event: MouseEvent|TouchEvent) => void)|undefined

    public onDown:((event: MouseEvent|TouchEvent) => void)|undefined
    private onUp:((event: MouseEvent|TouchEvent) => void)|undefined

    constructor(element: HTMLElement, delay: number) {
        this.element = element
        this.delay = delay
    }

    private animation(event: MouseEvent|TouchEvent) {
        this.startAnimation = true;
        this.elapse.point();

        const p = () => {
            let el = this.elapse.elapse()
            this.progress?.(Math.min(el, this.delay), this.delay);
            if (!this.startAnimation) {
                this.cancel?.(event as any)
            } else if (el >= this.delay) 
                this.longPressed?.(event as any)
            else 
                requestAnimationFrame(p);
            
        }
        requestAnimationFrame(p);
    }

    setup(): LongPressListener {
        this.cleanup();

        this.onDown = event => {
            // event.preventDefault();
            if (event instanceof MouseEvent && event.button !== 0) return; // 좌클릭만 처리
            this.element.addEventListener('mouseup', this.onUp as any)
            this.element.addEventListener('touchend', this.onUp as any)
            this.animation(event);
            this.start?.(event as any);
        }

        this.onUp = () => {
            this.startAnimation = false;
            this.element.removeEventListener('mouseup', this.onUp as any)
            this.element.removeEventListener('touchend', this.onUp as any)
        }

        this.element.addEventListener('mousedown', this.onDown as any)
        this.element.addEventListener('touchstart', this.onDown as any)

        return this;
    }

    cleanup(): LongPressListener {
        if (this.onDown) {
            this.element.removeEventListener('mousedown', this.onDown as any)
            this.element.removeEventListener('touchstart', this.onDown as any)
            this.onDown = undefined
        }

        if (this.onUp) {
            this.element.removeEventListener('mouseup', this.onUp as any)
            this.element.removeEventListener('touchend', this.onUp as any)
            this.onUp = undefined
        }


        return this;
    }

}