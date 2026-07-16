# GameAgentCreator 术语统一表

本表用于统一 Creator 中文界面里“刻意保留英文”的术语展示方式，避免同一概念在不同页面里出现纯英文、纯中文、半中半英混用的情况。

## 统一规则

- 中文界面里，保留英文术语时统一写成 `English（中文注解）`。
- API 字段名、关系类型、模式枚举等代码标识保持原样，不改大小写，不改下划线。
- 同一个术语在标签、提示、警告、说明文案里尽量使用同一套中文注解。

## 核心术语

| 原术语 | 统一展示 | 中文注解 / 使用说明 |
| --- | --- | --- |
| `Primary Parent` | `Primary Parent（主父节点）` | 大纲树中的稳定主层级父节点。 |
| `Relations` | `Relations（关系）` | 独立于大纲树的图关系集合。 |
| `External Parents` | `External Parents（外部父级）` | UI 里展示额外父向关系的集合标签。 |
| `located_at` | `located_at（当前位置关系）` | 表示节点当前位于哪个环境，不表示稳定归属。 |
| `belongs_to` | `belongs_to（稳定归属）` | 表示组织、阵营、资产或拥有关系。 |
| `subordinate` | `subordinate（隶属汇报）` | 表示指挥、汇报或控制链。 |
| `external_parent` | `external_parent（辅助父级作用域）` | 仅用于补充第二条父向作用域链，不进入默认上下文组装。 |
| `DAG` | `DAG（有向无环图）` | 用于说明非树状但仍有方向性的结构。 |
| `Publish Up` | `Publish Up（继续向上发布）` | 在既定作用域链上继续向更高层传播。 |
| `Pipeline Mode` | `Pipeline Mode（管线模式）` | 控制任务或传播处理方式的模式项。 |
| `Schema` | `Schema（结构定义）` | 用于描述结构、字段或能力输入输出格式。 |
| `Capability Schema` | `Capability Schema（能力结构定义）` | 能力项的结构定义，通常配合 JSON 对象使用。 |
| `Request ID` | `Request ID（请求 ID）` | 用于串联日志、轨迹和调试信息的请求标识。 |
| `LLM` | `LLM（大语言模型）` | 大模型能力相关术语统一保留缩写。 |
| `LLM Response` | `LLM Response（大语言模型响应）` | 原始模型返回内容。 |
| `Tick` | `Tick（时间刻）` | Engine 内部的最小时间推进单位。 |
| `World Tick` | `World Tick（世界时间刻）` | 面向世界级状态推进的 Tick 语义。 |
| `Tick Scale Mode` | `Tick Scale Mode（时间刻尺度模式）` | Tick 推进使用的尺度策略。 |
| `Tick Units` | `Tick Units（时间刻单位）` | Tick 时间体系中的单位列表。 |
| `Tick Min Unit` | `Tick Min Unit（最小时间刻单位）` | Tick 体系里最小的基础推进单位。 |
| `Tick Step` | `Tick Step（时间刻步长）` | 每次推进的最小单位步长。 |
| `Sub-Task DAG` | `Sub-Task DAG（子任务有向无环图）` | 子任务编排所使用的 DAG 结构。 |
| `World Time Settings` | `World Time Settings（世界时间设置）` | 世界时间推进相关的整体配置区。 |

## 维护建议

- 新增 Creator 文案时，如果术语已经在本表中，直接复用本表的统一展示。
- 如果必须新增一个保留英文的术语，优先补到本表后再落到界面文案中。
- 如果某个词只是普通功能文案，不需要保留英文时，优先直接使用自然中文，不要为了“看起来技术化”而强行夹英文。
