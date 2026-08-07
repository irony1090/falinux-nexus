import { Memoized } from '@/common/util/index.util';
import { EventInterface } from '@/common/util/lifecycle/event.util';

type LifecycleEntry = {
	init: () => void
	release: () => void
	isInit: boolean
}
type LifecycleHandler<E> = (el: E) => void;
type LifecycleHandlers<E> = {
	init: LifecycleHandler<E>
	release: LifecycleHandler<E>
}
type LifecycleEventMap<K> = {
	registered: [k: K, v: LifecycleEntry]
	unregistered: [k: K, v: LifecycleEntry]
	inited: [k: K]
	released: [k: K]
}

export class LifecycleRegistry<E> extends EventInterface<LifecycleEventMap<E>> {
	private memo: Memoized<E, LifecycleEntry>;
	private flag: boolean = false;

	constructor({init, release}: LifecycleHandlers<E>) {
		super()
		this.memo = new Memoized(el => {
			const initFunc = () => {
				ctx.isInit = true;
				init(el);
				this.emit('inited', el)
			}
			const releaseFunc = () => {
				ctx.isInit = false;
				release(el)
				this.emit('released', el)
			}
			const ctx = {
				init: initFunc,
				release: releaseFunc,
				isInit: false,
			}
			return ctx;
		})
	}

    values(): Array<E> {
        return this.memo.keys()
    }

	put(el: E) {
		const isNew = !this.memo.has(el);
		const ctx = this.memo.get(el);
		if (isNew && ctx) this.emit('registered', el, ctx);
	}

	remove(el: E) {
		const ctx = this.memo.get(el, false);
		if (!ctx) return;

		ctx.release();
		this.memo.remove(el);
		this.emit('unregistered', el, ctx);
	}

	isInit(): boolean {
		return this.flag;
	}

	initAll() {
		this.flag = true;
		for (const el of this.memo.keys()) {
			const ctx = this.memo.get(el)
			if (!ctx || ctx.isInit) continue;
			ctx.init();
		}
	}

	releaseAll() {
		this.flag = false;
		for (const el of this.memo.keys()) {
			const ctx = this.memo.get(el)
			if (!ctx || !ctx.isInit) continue;
			ctx.release();
		}
	}

}