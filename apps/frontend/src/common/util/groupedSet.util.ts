
export class GroupedSet<K, V> {
	private map = new Map<K, Set<V>>();

	put(key: K, value: V) {
		let set = this.map.get(key);
		if (!set) this.map.set(key, set = new Set());
		set.add(value);
	}

	remove(key: K, value: V) {
		const set = this.map.get(key);
		if (!set) return;
		set.delete(value);
		if (set.size === 0) this.map.delete(key);
	}

	has(key: K): boolean {
		return this.map.has(key);
	}

	forEach(key: K, fn: (value: V) => void) {
		this.map.get(key)?.forEach(fn);
	}
}
