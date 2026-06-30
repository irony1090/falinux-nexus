
export const DATE_TIME_FORMAT_SHORT = 'YY-MM-DD HH:mm:ss';
export const DATE_FORMAT = 'YYYY-MM-DD';
export const DATE_REG = /^\d{2,}-\d{1,2}-\d{1,2}$/;
export const TIME_STR = [
    'twelve',
    'one', 'two', 'three', 
    'four', 'five', 'six', 
    'seven', 'eight', 'nine', 
    'ten', 'eleven'
]

type Func<T> = (val: any) => T
type AddConstantFunc = Record<string, Func<any>>
type CreateOptions = Partial<{
    constants: boolean
}>
type Constant<T, K, Add extends AddConstantFunc> = {
    key: () => K,
    value: () => T;
    validation: () => boolean;
    init: (val: T) => void;
} & { [Key in keyof Add]: () => ReturnType<Add[Key]> }

export const createConstant = <T extends Record<string, any>, A extends Record<string, Func<any>>>(keyVal: T, addConstantFunc: A = {} as any) => {
    type Keys = keyof T;
    type AKeys = keyof A;
    
    const keys: (Keys)[] = [];
    const values: T[Keys][] = [] ;

    let flag = false;
    Object.entries(keyVal).forEach(([k, v]) => {
        if (flag) return;
        if (values.includes(v)) {
            flag = true
            return
        }
        keys.push(k as Keys);
        values.push(v);
    })
    if (flag) throw Error('Duplicate parameter values are not allowed')

    const create = <T>(val?: T, { constants: isConst = false }: CreateOptions = {}): Constant<T|undefined, Keys|undefined, A> => {
        let key: undefined|Keys  = undefined;
        let val_ = val;
        let validation: boolean = false
        const addFuncs: Record<AKeys, () => ReturnType<A[AKeys]>> = {} as any

        const init = (v: any) => {
            // if (constants) throw Error('this is constants')
            val_ = v;
            const idx = values.findIndex(v => v === val_);
            key = keys[idx];
            validation = idx > -1

            Object.entries(addConstantFunc).forEach(([k, func]) => {
                const v = func(val_);
                (addFuncs as any)[k] = () => v
            })
        }

        init(val);

        return {
            key: () => key,
            value: () => val_,
            validation: () => validation,
            init: isConst ? () => {throw Error('this is constants')} : init,
            ...addFuncs,
        }
    }

    const record = {}
    const constants = values.reduce((rcd, v, i) => {
        const cst = create(v, { constants: true });
        (rcd as any)[keys[i]] = cst;
        (record as any)[cst.value()!] = cst;
        return rcd;
    }, {} as { [K in Keys]: Constant<T[K], K, A>})

    type Cnst = typeof constants[keyof typeof constants]

    const constant = (val: any): Cnst | undefined => {
        const cst = (record as any)[val]
        return cst;
    }

    return {
        keys,
        values,
        raw: keyVal,
        constants,
        constant,
        create,
        _keyType: undefined as any as Keys,
        _valType: undefined as any as T[Keys],
        _constType: undefined as any as Cnst
    }
}


export const BOOLEAN_CONST = createConstant({
    TRUE: true as const,
    FALSE: false as const,
})