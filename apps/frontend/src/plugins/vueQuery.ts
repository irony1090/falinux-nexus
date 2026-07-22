import { QueryClient, VueQueryPlugin, type VueQueryPluginOptions } from '@tanstack/vue-query';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: 60 * 1000, // 1분
            retry: 2,
            refetchOnWindowFocus: false
        }
    }
})

export const vueQueryPluginOption: VueQueryPluginOptions = {
    queryClient
}




export default VueQueryPlugin;