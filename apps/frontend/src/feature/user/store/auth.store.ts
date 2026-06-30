import { inject, provide, readonly, ref, watch } from 'vue';
import { checkSession, signIn, signOut, type UserResponse } from '../api/user.api';
import { equals } from '@/common/util/index.util';
import { useRoute, useRouter } from 'vue-router';


const AUTH_STORE_KEY = 'AUTH_STORE_KEY';

export const provideAuthStore = () => {
    const router = useRouter();
    const route = useRoute();
    
    const auth = ref<UserResponse|null|undefined>()

    const loadding = ref(false);

    const confirm  = () => {
        loadding.value = true;
        return checkSession()
        .then(res => {
            auth.value = res;
        }).catch(_err => {
            auth.value = null;
        })
        .finally(() => loadding.value = false)
    }

    const logout = () =>{
        loadding.value = true;
        signOut().finally(confirm)
    }
    const login = (id: string, pw: string) => {
        loadding.value = true;
        return signIn({
            identification: id,
            password: pw
        }).then(async _ => {
            await confirm()
            // auth.value = res
            // return res;
        }).finally(() => {
            loadding.value = false;
        })
    }


    watch(auth, val => {
        if (loadding.value || val !== undefined)
            return;
        confirm()
    }, { immediate: true})
    
    const ctx = {
        auth: readonly(auth),
        loading: readonly(loadding),
        logout, login,
    }
    watch([auth, route], ([val, rVal]) => {
        if (!!val || !rVal) return;
        
        if (rVal.fullPath !== '/') {
            router.push('/login')
            // routerMng.login().move()
        }
    })

    provide(AUTH_STORE_KEY, ctx)
    return ctx;
}


type UseAuthStore = ReturnType<typeof provideAuthStore>

export const useAuthStore = () => {
    const context = inject<UseAuthStore>(AUTH_STORE_KEY)!;
    if (!context) throw new Error('useAuthStore is not provided');
    return context;
}
