import type { Dayjs } from 'dayjs';
import type { ElementType, Vec2 } from './index.type';
import dayjs from 'dayjs';

export const equals = (a:any, b:any): boolean => {
	if (isNil(a) || isNil(b)){
		// console.log('[EQ_1]', [a, b]);
		return a === b
	}else if (typeof a !== typeof b){
		// console.log('[EQ_2]', [a, b]);
		return false;
	}else if (typeof a === 'object' && typeof b === 'object') {
		if (a instanceof Date && b instanceof Date)
			return a.getTime() === b.getTime()
		else if (Array.isArray(a) && Array.isArray(b)) {
			// 배열 길이가 다르면 false
			if (a.length !== b.length) return false;
			// 모든 요소가 같아야 true
			return a.every((ae, i) => equals(ae, b[i]))
		} else {
			// 객체 키 개수가 다르면 false
			const keysA = Object.keys(a);
			const keysB = Object.keys(b);
			if (keysA.length !== keysB.length) return false;
			// 모든 키와 값이 같아야 true
			// console.log('[EQU] OBJ', {...a}, {...b})
			return keysA.every(k => Object.prototype.hasOwnProperty.call(b, k) && equals(a[k], b[k]))
		}
	} else {
		return a === b;
	}
}

export const removeUnit = (str: string, def: number = NaN) => {
	// const onlyNum = str.replaceAll(/[-+]\d*\.?\d+/gi, '');
	const unit = str.replace(/^[-+]?\d+(\.\d+)?/gi, '');

	const onlyNum = str.substring(0, str.length - unit.length)
	// console.log(str, onlyNum, unit);
	const result = onlyNum.length === 0 ? def : Number(onlyNum)
	return isNaN(result) ? def : result;
}

export const isNull = (val: any): val is null => val === null;
export const isUndefined = (val: any): val is undefined => val === undefined;
export const isNil = (val: any): val is undefined | null => isNull(val) || isUndefined(val);


export const extractElement = <T>(origin: T, select?: (p: T) => ElementType<T>): ElementType<T> => {
	if (Array.isArray(origin)) {
		return select ? select(origin) : origin[0]
	} else {
		return origin as ElementType<T>;
	}
}

export const dynamicUnit = (num: number, units: Array<string>, diff: number): [number, string] => {
	let value = num;
	let unitIndex = 0;

	while (value >= diff && unitIndex < units.length - 1) {
		value /= diff;
		unitIndex++;
	}

	return [
		Math.round(value * 100) / 100, // 소수점 2자리로 반올림
		units[unitIndex]!
	];
}

export type DecomposeUnitsProps = {
	max: number
	unit: string
}
type DecomposeUnits = {
	val: number
	unit: string
}
export const decomposeUnits = (num: number, units: Array<DecomposeUnitsProps>): Array<DecomposeUnits> => {
	let val = num;
	// const remain: Array<number> = [];
	const arr: Array<DecomposeUnits> = []
	for (let i = 0; i < units.length; i++) {
		const unit = units[i]!;
		if (val < unit.max) {
			arr.push({ val: val, unit: unit.unit})
			val = 0
			break;
		}
		const natureNum = Math.floor(val / unit.max)
		arr.push({ val: val % unit.max, unit: unit.unit})
		val = natureNum;
	}

	if (val !== 0) {
		const lastIndex = arr.length-1;
		arr[lastIndex]!.val += val * units[lastIndex]!.max
	}

	return arr.filter(({val}) => val !== 0).reverse();
}


type Validator<T> = (p: T) => void | string
export const createValidator = <T>(filter: Array<Validator<T>> | Validator<T>) => {

	const f = Array.isArray(filter) ? filter : [filter];

	return (val: T): string | void => {
		for (const isError of f) {
			const rst = isError(val);
			if (typeof rst === 'string' ) {
				return rst
			}
		}

		return
	}
}


export const toDayjs = (dateStr: string|Date|null|undefined): Dayjs | null => {
	if (dateStr) {
		const d = dayjs(dateStr);
		return d.isValid() ? d : null;
	} else {
		return null;
	}
}

export class TimeElapse {
	private curDate: number = Date.now();

	point(): TimeElapse {
		this.curDate = Date.now();
		return this;
	}

	elapse(): number {
		return Date.now() - this.curDate;
	}

}

export class Stale<K, V> {
	private stale = new Map<K, V>();

	init(map: Map<K, V>) {
		this.stale.clear();
		for (const [k, v] of map) this.stale.set(k, v);
	}
	touch(k: K) {
		if (this.stale.has(k)) this.stale.delete(k)
	}
	getStale(): Array<[K, V]> {
		const arr = [] as Array<[K, V]>;
		for (const entry of this.stale) arr.push(entry);
		return arr;
	}
}

export class Memoized<K, V> {
	private cache = new Map<K, V>();
	private extract: ((input: K) => V|undefined);

	private stale: Stale<K, V> | null = null;

	constructor(extract: (input: K) => V) {
		this.extract = extract;
	}

	getStaleAndInit(): Stale<K, V> {
		if (!this.stale) this.stale = new Stale()
			
		this.stale.init(this.cache);

		return this.stale;
	}

	get(input: K, extractIfNotExist: boolean = true) {
		if (this.cache.has(input)) {
			return this.cache.get(input)!;
		} else {
			const rst = extractIfNotExist ? this.extract(input) : undefined;
			if (rst !== undefined) this.cache.set(input, rst);
			return rst;
		}
	}

	remove(input: K) {
		const el = this.cache.get(input);
		if (el) this.cache.delete(input);
		return el;
	}

	clear() {
		this.cache.clear();
		this.stale?.init(this.cache);
	}

	size() {
		return this.cache.size;
	}
}

export const createMemoized = <I, O>(extract: (input: I) => O) => {
	const cache = new Map<I, O>();

	return {
		get: (input: I) => {
			if (cache.has(input)) {
				return cache.get(input)!;
			} else {
				const rst = extract(input);
				cache.set(input, rst);
				return rst;
			}
		},
		clear: () => cache.clear(),
		size: () => cache.size,
	}
}

export const overlaps = (a: Vec2, b: Vec2) => {
	const aMin = Math.min(...a);
	const aMax = Math.max(...a);
	const bMin = Math.min(...b);
	const bMax = Math.max(...b);

	return aMax >= bMin && bMax >= aMin;
}

// type DeepMapObj = {
// 	wrap: Object | Array<any>
// 	key: string|number
// 	val: any
// }
// export const deepMap = (obj: any, callback: (org: DeepMapObj, rst: DeepMapObj) => any) => {
// 	if (Array.isArray(obj)) {
// 		// const newArr = [];
		
// 		// return obj.map((v, i) => )
// 	}

// }


export class CircularBuffer<B> {
	private size: number;
	private buf: Array<B> = [];
	private head: number = 0;
	private tail: number = 0;
	private count: number = 0;
	private setHead: (beforebeforeTail: B, beforeTail:B, tail: B) => undefined | number
	
	// get head(): number {
	// 	return this.h
	// }
	// set head(val: number) {
	// 	this.h = val;
	// }

	// get tail(): number {
	// 	return this.t;
	// }
	// set tail(val: number) {
	// 	this.t = val;
	// }

	constructor(size: number = 8, setHead: (tail_2: B, tail_1: B, tail: B) => undefined|number = () => undefined) {
		this.size = size;
		this.setHead = setHead;
	}

	private getIndex(diff?: number): number {
		if (diff === undefined ) return this.head;
		else if (diff > 0) return (this.head + diff) % this.size;
		else if (diff < 0) return ((this.tail + diff) % this.size + this.size) % this.size;
		else return this.head;
	}

	append(val: B) {
		this.tail = (this.head + this.count) % this.size;
		this.buf[this.tail] = val;
		// this.print('START')
		if (this.count < this.size) this.count ++;
		else this.head = (this.head + 1) % this.size

		if (this.count > 2) {
			const flag = this.setHead(this.buf[this.getIndex(-2)]!, this.buf[this.getIndex(-1)]!, this.buf[this.tail]!)
			if (flag !== undefined && flag > -1) {
				const f = Math.min(flag, 2)
				this.head = ((this.tail + (f - 2)) % this.size + this.size) % this.size
				this.count = 3 - f;
				// this.head = (this.tail - 1 + Math.min(1, flag) ) % this.count;
			}
		}
		// this.print('E N D')

	}

	oldest(): B|undefined {
		return this.buf[this.head]
	}

	newest(): B|undefined {
		return this.buf[this.tail]
	}

	clean() {
		this.head = 0;
		this.count = 0;
		this.tail = 0;
		this.buf[this.head] = undefined as B
	}

	print(pre: string) {
		console.log(`(${pre}) [head]: ${this.head}, [count]: ${this.count}, [buf]:`, this.buf);
	}
}

// [LABEL] 유틸-숫자-소수
export const decimal = (num: number, digit: number, def: number): string => {
	const v = removeUnit(num.toString(), def)
	return v.toFixed(digit);
}