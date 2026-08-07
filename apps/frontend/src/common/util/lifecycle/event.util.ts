

export abstract class EventInterface<EventMap extends Record<string, Array<any>>> {
    protected event: Partial<{ [K in keyof EventMap]: (...args: EventMap[K]) => void }> = {}

    on<K extends keyof EventMap>(type: K, func: (...args: EventMap[K]) => void): () => void {
        this.event[type] = func;

        return () => {
            if (this.event[type] === func) this.event[type] = undefined;
        };
    }

    protected emit<K extends keyof EventMap>(type: K, ...args: EventMap[K]) {
        this.event[type]?.(...args);
    }
}