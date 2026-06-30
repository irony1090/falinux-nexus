import type { Diff, DiffField } from './index.type';
import { isUndefined } from './index.util';

type FieldMapper<T, V> = {                                                                                                  
	get: (el: T) => V                                                                                                       
	eq?: (a: V | undefined, b: V) => boolean|undefined  // 생략 시 기본 룰 적용
}
class ArrayUtil <T> {
	private arr: Array<T>;
	constructor(arr: Array<T>) {
		this.arr = arr;
	}

	diff<Shape extends Diff<Record<string, any>>>(
		map: {[K in keyof Shape]: FieldMapper<T, Shape[K]['current']>}
	): Array<Shape> {
		const ent = Object.entries(map);

		return this.arr.reduce((rst, el) => {
			const before = rst[rst.length-1];
			const newDiff = ent.reduce((newObj, [k, v]) => {
				const elVal = v.get(el);
				const isEquals = !v.eq 
					? (!isUndefined(before) ? elVal === before[k]?.current : undefined)
					: v.eq(before?.[k]?.current, elVal)

				newObj[k] = {
					before: before?.[k]?.current,
					current: elVal,
					isNotEquals: isUndefined(isEquals) ? undefined : !isEquals
				} as DiffField<any>
				return newObj;
			}, {} as any)

			rst.push(newDiff);
			return rst;
		}, [] as Array<Shape>)
	}
}
export const arrayUtil = <El>(arr: Array<El>) => new ArrayUtil(arr)