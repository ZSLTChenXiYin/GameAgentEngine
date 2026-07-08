# GameAgentCreator 指南

**中文** | [**English**](./GUIDE_GAMEAGENTCREATOR_EN.md)

GameAgentCreator 是 GameAgentEngine 附带的浏览器可视化编辑器。

---

## 打开 Creator

在浏览器中打开：

`tools/source/web/GameAgentCreator/index.html`

如果你的环境支持，也可以通过 CLI 的 `inspect` 流程进入。

---

## 主要界面布局

- 顶栏：世界选择、语言切换、主题切换、配置入口
- 左侧树：层级化节点大纲
- 中央区域：世界页、快照页、计划页、设置页、策略页、连续性页、日志页、轨迹页
- 右侧区域：节点摘要与挂载数据

---

## 当前支持的编辑流程

### 世界操作

- 创建世界
- 在世界页里修改当前世界名称
- 通过 `fork` 创建工作副本
- 保存存档快照
- 校验、恢复、删除快照
- 查看并审批待审核的世界变更计划
- 编辑世界设置
- 编辑世界策略

### 节点操作

- 创建节点
- 编辑节点
- 删除叶子节点
- 复制节点
- 从节点操作中添加“被指向关系”
- 将节点拖到另一个节点上，修改父子关系
- 将节点拖到根级放置区，清空父节点

当前节点复制会复制：

- 当前节点本身
- 开启子树复制时的整棵子树
- 挂载组件
- 挂载记忆
- 仍然完全处于复制子树内部的关系

“添加被指向关系”会打开与关系创建一致的关系编辑弹窗，用于从当前节点向其他节点创建任意关系类型，而不再局限于过去的外父节点场景。

### 组件编辑与校验

- Creator 会根据共享组件元数据提示当前组件是强类型、弱类型还是纯文本
- 编辑 `autonomous` 时，会按结构化配置校验关键字段
- 编辑 `world_state`、`story_state`、`story_history`、`tick_policy` 时，会按字段结构做连续性状态校验
- 编辑 `profile` 时，要求输入合法 JSON 对象
- 当前其他内置文本类组件允许直接按文本编辑

### 运行时操作

- 推进 Tick
- 触发自主行为
- 评估事件影响
- 局部范围推进
- 时间线重规划

### 记忆传播

- 从节点详情中的记忆列表显式触发传播
- 支持选择 `upward`、`tag_broadcast`、`targeted`、`manual` 四种模式
- 支持填写标签、目标节点、最大深度和 `publish_up`

### 可观测性

- 连续性聚合页，用于查看最近一次 `world_tick` 的整体连续性工件
- `Continuity Diff` 卡片，用于对比当前 tick 与上一轮 tick 的摘要和事实变化
- 推理日志
- 调试轨迹
- configured / effective pipeline mode 可见性
- round usage 可见性

## 连续性工作流

- 打开 `Continuity` 页面，先加载最近一次 world-oriented continuity bundle
- 使用 request 过滤器，把日志和轨迹收敛到单个 `request_id`
- 在 `Continuity Diff` 中对比 `Latest Tick Summary`、`Previous Tick Summary` 和事实增删
- 需要直接修连续性状态时，再切换到 `State` 页面编辑 `world_state`、`story_state`、`story_history`、`tick_policy`
- 除非你明确要重建引擎生成状态，否则保持 `state_snapshot` 只读

### 连续性状态与时间线

- `State` 页面用于查看 `world_state`、`story_state`、`story_history`、`tick_policy`、`state_snapshot`
- `Timelines` 页面用于查看最近的 tick 历史归档与结构化 payload
- `State` 页面可直接编辑除 `state_snapshot` 外的连续性状态组件
- `state_snapshot` 保持只读，更适合作为引擎生成的检查点观察面

当你要排查剧情上下文失真时，推荐按以下顺序看：

1. `Timelines` 查看最近一次 tick 的 `reply` / `future_outline`
2. `State` 查看 `world_state.canonical_facts` 和 `story_history.entries`
3. `Logs` 查看 request / response / detail
4. `Traces` 查看 Debug 模式下的 prompt 与解析结果

---

## 快照页说明

`Snapshots` 页面有两种常见视图：

- 当选中的是普通可运行世界时，页面展示该世界创建过的存档快照列表
- 当选中的是存档快照世界时，页面会展示当前快照的元数据，并列出其源世界上的全部存档快照

---

## 当前约束

- Creator 通过 HTTP 与引擎通信，因此依赖正在运行的服务端
- 世界重命名与节点复制依赖新版 API 路由
- Creator 的组件校验提示来自打包时生成的 `js/component-meta.js`
- 用于打包与分发的 Creator 副本，是 `tools/source/web/GameAgentCreator`
