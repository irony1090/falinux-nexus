import type { AxiosError, AxiosResponse } from 'axios';
import axios from 'axios';
import { readTabId, TAB_ID_HEADER } from '../util/tabId.util';

// type ResultStatus = 'success' | 'fail';
export type CommonResponse<T = {}> = {
    status: 'success' | 'fail';
} & T

export type CatchErrorResponse = {
    message: string;
    status?: number;    // HTTP 상태 — 서버 응답이 있을 때만
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
            throw { message, status: axiosError.response.status } as CatchErrorResponse;
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

// 요청 탭 식별(크기 우선권 등) — 서버가 요청 세션의 탭인지 확인한다
BaseAxios.interceptors.request.use(config => {
    const tabId = readTabId();
    if (tabId) config.headers.set(TAB_ID_HEADER, tabId);
    return config;
});

// 401(세션 만료) 알림 — common이 feature/user를 import하지 않도록 등록 방식으로 둔다 — 로그인-3(401만 로그아웃)
// `/users/session`은 auth가 직접 결과를 처리하므로 알리지 않는다
const unauthorizedListeners: Array<() => void> = [];
export const onUnauthorized = (cb: () => void) => unauthorizedListeners.push(cb);

BaseAxios.interceptors.response.use(undefined, (error: AxiosError) => {
    if (error.response?.status === 401 && !error.config?.url?.startsWith('/users/session'))
        unauthorizedListeners.forEach(cb => cb());
    return Promise.reject(error);
});
