
type Test<P> = (val: P) => true | string;
type Tests<T> = {
    [K in keyof T]: Test<T[K]> | undefined
}
export class EntityValidator<T extends {}> {
    
    private tests: Tests<T>
    public messages: Array<string> = [];

    constructor(tests: Tests<T>) {
        this.tests = tests;
    }

    setup(param: Partial<T>):T|null {
        type K = keyof T;

        this.cleanMessages();

        const r = Object.keys(this.tests).reduce((rst, k_) => {
            const k = k_ as K;
            const test = this.tests[k];
            const val = param[k] as T[K];
            if (test) {
                const flag = test(val);
                if (flag === true) 
                    rst[k] = val
                else if (typeof flag === 'string')
                    this.messages.push(flag);
                    
            } else
                rst[k] = val

            return rst;
        }, {} as T)

        return this.messages.length > 0 ? null : r;
    }

    private cleanMessages() {
        if (this.messages.length > 0)
            this.messages.splice(0, this.messages.length);
    }

}