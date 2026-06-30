export type Vec2<T = number> = [T, T];
export type Vec3<T = number> = [T, T, T];
export type Vec4<T = number> = [T, T, T, T];

export type SizeGrade = 'XS' | 'SM' | 'MD' | 'LG' | 'XL'

export type Request = {
	Status: 'WAITING' | 'PENDING',
	Result: 'SUCCESS' | "FAIL"
};

export type Connect = {
	Status: 'PENDING' | 'CONNECTING' | 'CONNECTED' | 'FAILED'
}

// 깊은 필드까지 옵셔널
export type DeepPartial<T> = T |
	(
		T extends Array<infer U>
		? DeepPartial<U>[]
		: T extends Map<infer K, infer V>
		? Map<DeepPartial<K>, DeepPartial<V>>
		: T extends Set<infer M>
		? Set<DeepPartial<M>>
		: T extends object
		? { [K in keyof T]?: DeepPartial<T[K]>; }
		: T
	);

export type EnumType<T extends string | number> = { [K in T]: K };

/**오브젝트(O) 내부의 필드의 타입을 교체(R)*/
export type Replace<O, R> = Omit<O, keyof R> & R;

// 메소드에 해당하는 값만 가지고 오기
export type MethodsOf<T> = {
	[K in keyof T]: T[K] extends Function ? K : never
}[keyof T];

export type ElementType<T> = T extends  Array<infer E> ? E : T;



// 빈 값 타입
export type EmptyType = 'null' | 'undefined';

// 단순 값 타입
export type ValueType = 'bigint' | 'boolean' | 'function' | 'number' | 'string' | 'symbol' | EmptyType;

// 컨테이너 타입
export type ContainerType = 'object' | 'array' | 'tuple';

type Allow = '*';
// 모든 타입
export type TypeOf = ValueType | ContainerType | Allow;

type DeclarationType = ContainerType | Allow;

type Declaration = {
	type: DeclarationType | Array<DeclarationType | EmptyType>;
	structure: Structure | Array<ValueType | Declaration>
}

export interface Structure {
	[key: string]: ValueType | ValueType[] | Declaration
}

// export type Diff<V> = {
// 	before: V|undefined
// 	current: V
//     isNotEquals: boolean | undefined
// }

export type DiffField<V> = {
	before: V | undefined
	current: V
	isNotEquals: boolean | undefined
}

export type Diff<Shape extends Record<string, any>> = {                                                                                        
	[K in keyof Shape]: DiffField<Shape[K]>
}  