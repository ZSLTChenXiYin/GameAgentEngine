# Demo 世界：灰港边境

**中文** | [**English**](./DEMO_WORLD_GRAY_HARBOR_EN.md)

「灰港边境」是一个演示用的边境治理模拟世界，用于展示 GameAgentEngine v0.2.0 的核心功能。

---

## 世界概览

灰港边境是一个资源紧张、派系林立的边境城镇。玩家作为议会执政官，需要在各方势力之间周旋，应对突发事件，维持边境稳定。

### 节点结构

```
灰港边境 (world)
├── 议事厅 (location)
│   └── 圆桌议室 (location)
│       └── 灰港议会 (faction) — 执政派系
│           ├── 艾琳议长 (npc) — 冷静、克制、善于权衡
│           └── 布莱姆总管 (npc) — 注重资源实数，反感空谈
├── 雾湾矿场 (location)
│   └── 雾湾征购站 (location)
│       └── 铁潮商会·雾湾矿场分部 (faction)
├── 北门要塞 (location)
│   └── 北门军营 (location)
│       └── 北境守军 (faction)
│           └── 赛洛指挥官 (npc) — 务实、重战备、信奉实力
├── 河岸集市 (location)
│   └── 河岸货栈 (location)
│       └── 铁潮商会 (faction)
│           └── 米拉代表 (npc) — 精明、圆滑、以利益为导向
```

### 角色资源状态

世界通过自定义组件跟踪各区域的资源：

- `resource_state`（世界级）：food=62, order=58, defense=49, morale=55, treasury=46
- `district_state`（地区级）：稳定性、压力值、产出

---

## 导入方式

```bash
# 使用 DevCli 导入
GameAgentDevCli import demo-world.yaml --reset
```

`--reset` 参数会在导入前清空数据库。如果不加 `--reset`，Demo 世界会在现有数据库中追加。
