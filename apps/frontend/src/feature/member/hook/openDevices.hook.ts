import { computed } from 'vue'
import { useMemberDeviceList } from '../api/device.api'
import { useMemberBoard } from '../store/memberBoard.store'
import { kindOf, mainKeyOf } from '../util/memberGrid.util'

// 판에 열린 장비 타일들(타일 순서) + 접속 중 instance가 하나라도 있는지 — REF-device-move.md 이동-Q1(대상)
export const useOpenDevices = () => {
    const { tiles, admin } = useMemberBoard()
    const mainKeys = computed(() => tiles.value.filter(t => kindOf(t.key) === 'device').map(t => mainKeyOf(t.key)))
    const devices = useMemberDeviceList(mainKeys, admin)
    const anyOnline = computed(() => devices.value.some(d => d.instances.some(i => i.online)))
    return { devices, anyOnline }
}
