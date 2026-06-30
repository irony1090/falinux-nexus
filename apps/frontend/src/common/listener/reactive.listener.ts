
export class WatchedValue<T> {
    private target_: T
    public onRead: ((v: T, t_: WatchedValue<T>) => void) | undefined
    public onWrite: ((v: T, t_: WatchedValue<T>) => void) | undefined

    constructor(target: T) {
        this.target_ = target;
    }

    get target() {
        this.onRead?.(this.target_, this);
        return this.target_;
    }

    set target(v: T) {
        this.target_ = v;
        this.onWrite?.(v, this);
    }
}
