import type { SizeGrade } from '../util/index.type';

export type SizeConfig = {
    grade: SizeGrade,
    width: number, // 최대 사이즈
    baseFontSize: number // 기본 폰트 사이즈
}

export const SIZE_RECORD = {
    XS: { // 폴더블(접힘)
        grade: 'XS',
        width: 296,
        baseFontSize: 14
    },
    SM: { // 일반 핸드폰
        grade: 'SM',
        width: 420,
        baseFontSize: 14
    },
    MD: { // 태블릿
        grade: 'MD',
        width: 756,
        baseFontSize: 16
    },
    LG: { // 노트북
        grade: 'LG',
        width: 1280,
        baseFontSize: 16
    },
    XL: { // PC
        grade: 'XL',
        width: 1980,
        baseFontSize: 16
    }
} as const

export const getSizeConfig = (length: number): SizeConfig => {
    
    if (length <= SIZE_RECORD.XS.width) {
        return SIZE_RECORD.XS;
    } else if (length <= SIZE_RECORD.SM.width) {
        return SIZE_RECORD.SM;
    } else if (length <= SIZE_RECORD.MD.width) {
        return SIZE_RECORD.MD;
    } else if (length <= SIZE_RECORD.LG.width) {
        return SIZE_RECORD.LG;
    } else {
        return SIZE_RECORD.XL;
    }
    
}
type CompareType = 'UP'|'DOWN'
// export enum CompareType {
//     UP, DOWN
// }
export const compareSizeGrade = (direction: CompareType, target: SizeConfig, base: SizeConfig): boolean => {
    if (direction === 'UP') {
        return target.width >= base.width;
    } else {
        return target.width <= base.width;
    }
};