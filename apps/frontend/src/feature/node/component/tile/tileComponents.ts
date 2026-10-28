import type { TileType } from '@/feature/widget/util/tile.type'
import type { Component } from 'vue'
import DeviceTile from './DeviceTile.vue'
import FolderTile from './FolderTile.vue'
import TerminalTile from './TerminalTile.vue'
import ViewerTile from './ViewerTile.vue'

// 타일 종류 -> 그리는 컴포넌트(TileFrame을 채움). 새 종류 = 줄 하나 + TILE_KINDS 줄 하나 (REF-tile-kinds.md)
export const TILE_COMPONENTS: Record<TileType, Component> = {
    folder: FolderTile,
    terminal: TerminalTile,
    device: DeviceTile,
    viewer: ViewerTile,
}
