import { EventInterface } from '../util/lifecycle/event.util';

export type DragPoint = { x: number; y: number };

export type DragSessionOptions = {
    threshold: number   // 마우스 = 이만큼 움직이면 시작(px)
    holdMs: number      // 터치·펜 = 이만큼 누르고 있으면 시작
    holdSlop: number    // 터치·펜 = 시작 전에 이만큼 움직이면 취소(스크롤로 봄)
};

type DragEvents = {
    hold: [DragPoint]       // 터치 누르기 시작 — 시작 대기 표시용
    start: [DragPoint]
    move: [DragPoint]
    drop: [DragPoint]
    cancel: []              // 시작 전 취소 포함
};

export type DragState = 'idle' | 'pending' | 'dragging';

const DEFAULTS: DragSessionOptions = { threshold: 6, holdMs: 500, holdSlop: 10 };
const EAT_MS = 400;     // 끝난 뒤 이 안의 click·contextmenu는 버린다

// 끌기 시작 판정 + 추적 — 마우스 = 거리, 터치·펜 = 길게 누르기. 놓을 자리 계산은 소비처가 한다(REF-util-drag.md)
// 쓰는 법: 소비처가 pointerdown에서 대상이면 down(e) -> start/move/drop/cancel 이벤트. 터치 스크롤을 막으려면 blockTouch(root)
export class DragSession extends EventInterface<DragEvents> {
    private opt: DragSessionOptions;
    private state_: DragState = 'idle';
    private id = -1;
    private from: DragPoint = { x: 0, y: 0 };
    private touch = false;
    private timer: ReturnType<typeof setTimeout> | undefined;
    private eatUntil = 0;
    private origin: EventTarget | null = null;

    constructor(opt: Partial<DragSessionOptions> = {}) {
        super();
        this.opt = { ...DEFAULTS, ...opt };
    }

    get state() {
        return this.state_;
    }

    // 시작 대기. 이미 진행 중이거나 주 포인터·왼쪽 버튼이 아니면 false
    down(e: PointerEvent): boolean {
        if (this.state_ !== 'idle' || !e.isPrimary || (e.pointerType === 'mouse' && e.button !== 0)) return false;
        this.state_ = 'pending';
        this.id = e.pointerId;
        this.from = { x: e.clientX, y: e.clientY };
        this.touch = e.pointerType !== 'mouse';
        this.listen(true);
        if (this.touch) {
            // 손가락을 댄 요소 자체에도 막기 — 끄는 중 그 요소가 다시 마운트돼 문서에서 빠지면 touchmove는 그 요소로만 가고 root까지 안 온다(10-29 폰 실측)
            this.origin = e.target;
            this.origin?.addEventListener('touchmove', this.onTouchMove, { passive: false });
            this.timer = setTimeout(() => this.begin(this.from), this.opt.holdMs);
            this.emit('hold', this.from);
        }
        return true;
    }

    cancel() {
        if (this.state_ === 'idle') return;
        this.finish();
        this.emit('cancel');
    }

    // 끄는 동안 root 안 터치 스크롤을 막는다. 터치 시작 전부터 비수동 리스너가 있어야 해서 따로 건다 — 해제 함수 반환
    blockTouch(root: HTMLElement): () => void {
        root.addEventListener('touchmove', this.onTouchMove, { passive: false });
        return () => root.removeEventListener('touchmove', this.onTouchMove);
    }

    private onTouchMove = (e: Event) => {
        if (this.state_ === 'dragging' && e.cancelable) e.preventDefault();
    }

    private begin(p: DragPoint) {
        if (this.state_ !== 'pending') return;
        clearTimeout(this.timer);
        this.state_ = 'dragging';
        window.getSelection()?.removeAllRanges();
        this.emit('start', p);
    }

    private finish() {
        clearTimeout(this.timer);
        if (this.state_ === 'dragging') this.eatUntil = performance.now() + EAT_MS;
        this.state_ = 'idle';
        this.id = -1;
        this.origin?.removeEventListener('touchmove', this.onTouchMove);
        this.origin = null;
        this.listen(false);
    }

    private onMove = (e: PointerEvent) => {
        if (e.pointerId !== this.id) return;
        const p = { x: e.clientX, y: e.clientY };
        if (this.state_ === 'pending') {
            const dist = Math.hypot(p.x - this.from.x, p.y - this.from.y);
            if (this.touch) {
                if (dist > this.opt.holdSlop) this.cancel();
                return;
            }
            if (dist < this.opt.threshold) return;
            this.begin(p);
        }
        this.emit('move', p);
    }

    private onUp = (e: PointerEvent) => {
        if (e.pointerId !== this.id) return;
        if (this.state_ !== 'dragging') return this.cancel();
        this.finish();
        this.emit('drop', { x: e.clientX, y: e.clientY });
    }

    // 브라우저가 스크롤로 가져감·두 번째 손가락(펼치기 등) = 취소
    private onCancel = (e: PointerEvent) => {
        if (e.pointerId === this.id) this.cancel();
    }
    private onOtherDown = (e: PointerEvent) => {
        if (e.pointerId !== this.id) this.cancel();
    }
    private onKey = (e: KeyboardEvent) => {
        if (e.key !== 'Escape' || this.state_ !== 'dragging') return;
        e.preventDefault();
        e.stopPropagation();
        this.cancel();
    }
    private onBlur = () => this.cancel();

    // 누르는 동안의 기본 메뉴(길게 누르기) + 끈 직후 손 뗀 자리의 click·메뉴
    private onEat = (e: Event) => {
        const busy = this.state_ !== 'idle' && e.type === 'contextmenu';
        if (!busy && performance.now() > this.eatUntil) return;
        e.preventDefault();
        e.stopPropagation();
    }

    private listen(on: boolean) {
        const opt = { capture: true };
        if (on) {
            window.addEventListener('pointermove', this.onMove, opt);
            window.addEventListener('pointerup', this.onUp, opt);
            window.addEventListener('pointercancel', this.onCancel, opt);
            window.addEventListener('pointerdown', this.onOtherDown, opt);
            window.addEventListener('keydown', this.onKey, opt);
            window.addEventListener('blur', this.onBlur);
        } else {
            window.removeEventListener('pointermove', this.onMove, opt);
            window.removeEventListener('pointerup', this.onUp, opt);
            window.removeEventListener('pointercancel', this.onCancel, opt);
            window.removeEventListener('pointerdown', this.onOtherDown, opt);
            window.removeEventListener('keydown', this.onKey, opt);
            window.removeEventListener('blur', this.onBlur);
        }
        // 먹는 리스너는 끝난 뒤에도 잠깐 일해야 해서 처음 down에서 한 번 걸고 dispose까지 둔다
        if (on && !this.eating) {
            this.eating = true;
            window.addEventListener('contextmenu', this.onEat, { capture: true });
            window.addEventListener('click', this.onEat, { capture: true });
        }
    }
    private eating = false;

    // 소비처가 내려갈 때
    dispose() {
        this.cancel();
        window.removeEventListener('contextmenu', this.onEat, { capture: true });
        window.removeEventListener('click', this.onEat, { capture: true });
        this.eating = false;
    }
}
