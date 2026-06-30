<script setup lang="ts">
import { LazyListener } from '@/common/listener/lazy.listener';
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { ref, watch } from 'vue';
import type { SubmitEventPromise } from 'vuetify';
import { VCard, VCardTitle, VCardActions, VBtn, VForm, VTextField, VCheckbox, VImg } from 'vuetify/components'
import Cookies from 'js-cookie';
import { isNil } from '@/common/util/index.util';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { useRouter } from 'vue-router';

const REMEMBER_ID_COOKIE_KEY = 'remember_id';

const { loading, login, auth } = useAuthStore();
const isLazyLoading = ref(false);
const lazyListener = new LazyListener(false, { time: 750, startLazy: v => !v})
lazyListener.onValue = v => isLazyLoading.value = v;

const { openDialog } = useAppDialog();

const router = useRouter();

const id = ref(Cookies.get(REMEMBER_ID_COOKIE_KEY) ?? 'irony')
const pw = ref('!Fa1289')

const rememberId = ref(!isNil(Cookies.get(REMEMBER_ID_COOKIE_KEY)));


const idMsg = ref<string>()
const pwMsg = ref<string>()
const idValid = (): string | null => {
    if (id.value === '') return 'ID 입력이 필수입니다';
    return null;
}
const pwValid = (): string | null => {
    if (pw.value === '') return 'PW 입력이 필수입니다';
    return null;
}

const onSubmit = (e: SubmitEventPromise) => {
    e.preventDefault()
    const idM = idValid()
    if (idM) {
        idMsg.value = idM
    }

    const pwM = pwValid()
    if (pwM) {
        pwMsg.value = pwM
    }

    const exit = !!idM || !!pwM
    if (exit) return
    
    // routerMng.main().move()
    // router.push('/')

    login(id.value, pw.value)
    .then( _ => {
        if (rememberId.value)
            Cookies.set(REMEMBER_ID_COOKIE_KEY, id.value, { expires: 7 })
        else 
            Cookies.remove(REMEMBER_ID_COOKIE_KEY)
        // router.replace('/workspace')
        // console.log('[SUCCESS]', res);
    }).catch(err => {
        openDialog({
            type: 'error',
            content: err.message
        })
    })
}

const onRegister = () => {
    
}

watch(loading, val => {
    lazyListener.setValue(val);
}, {immediate: true})

watch(id, () => idMsg.value = undefined)
watch(pw, () => pwMsg.value = undefined)

watch(auth, val => {
    if (!val) return;
    router.push('/')
}, { immediate: true })

</script>

<template>
<div id="Login" ref="loginRef">
    <v-card class="card" ref="vRef" 
        variant="text"
        :loading="isLazyLoading" 
    >
        <div class="logo">
            <v-img 
                width="96"
                aspect-ratio="1"
                cover
                src="/favicon.ico"
            />
        </div>
        <v-card-title><h3 class="text-primary">NEXUS-supervisor</h3></v-card-title>
        <v-form @submit="onSubmit">
            <div class="content">
                <v-text-field 
                    label="ID"
                    name="idenfitication"
                    variant="outlined"
                    :messages="idMsg"
                    :error="!!idMsg"
                    autocomplete="off"
                    :disabled="isLazyLoading"
                    color="primary"
                    v-model="id"
                />
                <v-text-field 
                    label="PW" 
                    name="password"
                    type="password"
                    variant="outlined"
                    :messages="pwMsg"
                    :error="!!pwMsg"
                    autocomplete="off"
                    :disabled="isLazyLoading"
                    color="primary"
                    v-model="pw"
                />
                <v-checkbox 
                    density="compact"
                    hide-details
                    color="primary"
                    label="아이디 저장"
                    v-model="rememberId"
                />
            </div>
            <v-card-actions class="action">
                <v-btn
                    variant="flat"
                    color="primary"
                    type="submit"
                    :disabled="isLazyLoading"
                >
                    로그인
                </v-btn>
                <v-btn
                    variant="text"
                    color="primary"
                    :disabled="isLazyLoading"
                    @click="onRegister"
                >
                    회원가입
                </v-btn>
            </v-card-actions>
        </v-form>
    </v-card>
</div>
</template>

<style scoped lang="scss">

#Login {
    width: 100%;
    height: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
}
.card {
    width: 100%;
    max-width: 296px;
    margin: vars.$spacing-lg;
    .logo {
        width: 100%;
        .v-img {
            margin: 0 auto;
        }
    }
    .v-card-title {
        text-align: center;
    }
    .content {
        padding: vars.$spacing-lg;
        padding-bottom: 0;
        display: flex;
        flex-direction: column;
        justify-content: center;
        gap: vars.$spacing-md;
        .v-checkbox {
            & :deep(.v-selection-control) {
                --v-input-control-height: 28px;
            }
            &:deep(.v-input__control .v-selection-control) {
                gap: vars.$spacing-md;
            }
        }
    }
    .action {
        padding-top: vars.spacing(6);
        display: flex;
        flex-direction: column;
        align-items: stretch;
        gap: vars.$spacing-md;
    }
}
</style>