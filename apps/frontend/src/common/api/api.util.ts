import type { AxiosError, AxiosResponse } from 'axios';
import axios from 'axios';

// type ResultStatus = 'success' | 'fail';
export type CommonResponse<T = {}> = {
    status: 'success' | 'fail';
} & T

export type CatchErrorResponse = {
    message: string;
    stack: string | undefined;
}
export const throwThen = <T>(response: AxiosResponse<CommonResponse>): T => {
    if (response.status === 200) 
        return  response.data as T;
    if (response.status !== 200) {
        console.log('[throwThen1]', response)
        throw response.data;
    } else if (response.data && 'status' in response.data && response.data?.status !== 'success') {
        console.log('[throwThen2]', response)
        throw response.data;
    }
    // const {status: _, ...other} = response.data;
    return response.data as T;
}

export const throwCatch = (error: AxiosError | CommonResponse) => {
    // console.log('[throwCatch]', error);
    if (error && 'isAxiosError' in error && error.isAxiosError) {
        const axiosError = error as AxiosError<CommonResponse<Record<'result' | 'message' | 'error', string>>>; 
        if (axiosError.response) {
            const { data } = axiosError.response || {};
            const message = (data?.result || data.message || data.error) ?? '서버 에러가 발생했습니다';
            throw { message } as CatchErrorResponse;
        } else {
            console.log('[THROW] isAxiosError', error);
            throw { 
                message: error.message,
                stack: (error as any).stack
            };
        }
    } else {
        if (error && 'status' in error) {
            throw { message: '[ERROR]' };
        }
        console.log('[THROW] ELSE', error);

        throw { };
    }
};


const getBaseURL = (): string => {
    const rst = !!import.meta.env.VITE_API_URL ? import.meta.env.VITE_API_URL : `${location.protocol}//${location.host}`
    return rst;
}

export const BaseAxios = axios.create({
    baseURL: getBaseURL(),
    timeout: 7500,
    withCredentials: true,
    headers: {
        'Content-Type': 'application/json',
    }
});
