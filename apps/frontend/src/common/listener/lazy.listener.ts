
// export const class LazyListener 

import type { Vec2 } from '../util/index.type';
import { isNil, TimeElapse } from '../util/index.util';
import { WatchedValue } from './reactive.listener';

export type LazyListenerOption<T> = Partial<{
    time: number
    startLazy: (v: T) => boolean
}>

export class LazyProcessListener<T> {
    private val: T;
    private time: number;
    private startLazy: ((v: T) => boolean)
    private isStart_ = false;
    private elapse = new TimeElapse()
    private process: WatchedValue<number>
    public onValue: ((v: T) => void) | undefined

    constructor(v: T, { time = 500, startLazy = () => true}: LazyListenerOption<T> = {}) {
        this.val = v;
        this.startLazy = startLazy;
        this.time = time;
        this.process = new WatchedValue(0);
        this.elapse.point();
    }

    set onProcess(listener: ((vec2: Vec2) => void) | undefined) {
        if (listener)
            this.process.onWrite = v => listener([v, this.time])
        else
            this.process.onWrite = undefined
    }

    get start() {
        return this.isStart_;
    }
    set start(v: boolean) {
        this.isStart_ = v;

        if (v) this.process.target = this.elapse.point().elapse();
    }

    private animation() {
        if (!this.start) return;
        const el = this.elapse.elapse();
        const processVal = Math.min(el, this.time);
        this.process.target = processVal;
        if (el >= this.time) {
            this.onValue?.(this.val);
            this.start = false;
        } else
            requestAnimationFrame(this.animation.bind(this));
        
    }

    setValue(v: T) {
        this.val = v;
        const flag = this.startLazy(v);

        if (flag) {
            const isStarted = this.start;
            this.start = true;
            if (!isStarted) {
                requestAnimationFrame(this.animation.bind(this));
                // this.onValue?.(this.val);
            }
            
        } else if (!this.start) 
            this.onValue?.(this.val);
        

    }
}
export class LazyListener<T> {
    private val: T
    private time: number;
    private lazyId_: number|undefined;
    private startLazy: ((v: T) => boolean)
    public onValue: ((v: T) => void) | undefined

    constructor(v: T, { time = 500, startLazy = () => true}: LazyListenerOption<T> = {}) {
        this.val = v;
        this.startLazy = startLazy;
        this.time = time;
    }

    get start() {
        return !isNil(this.lazyId_) && !isNaN(this.lazyId_);
    }


    set lazyId(id: number|undefined) {

        if (this.start) clearTimeout(this.lazyId_);
        
        this.lazyId_ = id;
    }

    setValue(v: T) {
        this.val = v;
        const flag = this.startLazy(v);
        if (flag) {
            // const isStarted = this.start;
            this.lazyId = setTimeout(() => {
                this.onValue?.(this.val);
                this.lazyId = undefined;
            }, this.time)

            // if (!isStarted) this.onValue?.(this.val);
        } else {
            if (!this.start) this.onValue?.(this.val);
        }

    }
}