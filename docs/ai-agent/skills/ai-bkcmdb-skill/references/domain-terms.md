# bk-cmdb 术语与字段映射

本文件约定 bk-cmdb 查询中用到的术语、需求字段到 CMDB 字段的映射，以及禁止编造的字段。

## 术语

- **业务（biz）** — CMDB 的顶层资源归属单元，`bk_biz_id` 唯一标识，`bk_biz_name` 为名称。资源池为特殊业务。
- **集群（set）** — 业务下的一层拓扑，`bk_set_id` / `bk_set_name`。
- **模块（module）** — 集群下的一层拓扑，`bk_module_id` / `bk_module_name`。主机最终挂在模块上。
- **拓扑路径** — 主机在业务内的归属链：`业务名 / 集群名 / 模块名`。
- **管控区（cloud area）** — 网络隔离区，`bk_cloud_id`（直连区为 0）。同一 IP 在不同管控区可能重复，定位主机时需结合管控区。
- **纳管** — 主机已录入 CMDB 并归属到某业务/模块。未纳管的 IP 查询会「检索为 0」。
- **租户** — 数据范围以调用方所在租户为准，每个租户有各自的业务、主机池、主机等资源。不从 Prompt 或用户话术读取租户。
- **资源池 / 主机池** — 当前租户下的特殊业务，用于管理**尚未分配到业务**的主机，导入的新主机默认会进入该租户的主机池。判定规则见下文「主机池识别」。
- **固资号** — 固定资产编号，`bk_asset_id`；**无固定格式**（可含字母/数字/符号，也可能是纯数字），不能凭前缀判定；可作为查主机的入口标识符之一。
- **主机标识符（查询入口）** — 定位主机的三类入口：内网 IP（`bk_host_innerip`）/ 主机 ID（`bk_host_id`，CMDB 内部纯数字）/ 固资号（`bk_asset_id`，无固定格式）。三者仅过滤字段不同，归属解析与输出视图一致；未指明类型时用 `OR` 同时匹配候选字段。
- **CMDB 记录更新时间** — 主机在 CMDB 中最近一次变更时间，来自 `open_list_hosts_without_biz` 返回的 `last_time` 或 `bk_updated_at`；用于判断数据在 CMDB 侧是否陈旧。
- **集群模板（set template）** — 用于标准化创建集群/模块结构的模板，使用`set_template_id` 指定。按模板创建集群时，系统会自动创建模板关联模块，不需要手动创建模块。
- **服务模板（service template）** — 模块标准配置模板，包含服务分类与进程模板等定义，按服务模板创建模块时，模块名称、服务分类均以模板为准。
- **服务分类（service category）** — 服务模板所属分类，使用`service_category_id` 指定。
- **进程模板（process template）** — 服务模板下定义进程实例默认参数的模板，使用`process_template_id` 指定。

## 主机池识别

主机池业务 ID 不从 Prompt 读取，也不写死。需要判定主机是否在主机池时，调用 internal MCP 上的 `get_resource_pool_biz`，取返回的 `data.bk_biz_id`，这就是**调用方所在租户**的主机池业务 ID。

- 查询到主机的 `bk_biz_id` 等于该值时，判定归属为主机池，直接显示为「主机池」，跳过 `open_search_business`（不查业务名）。
- 接口无请求体，不要传入租户 ID 或业务 ID。租户由调用方身份决定。
- 同一轮查询复用这一次结果。接口失败或未返回 `bk_biz_id` 时，按失败说明，不要改用猜测的 ID。
- 该工具只在 internal MCP 中注册，不对外暴露。按工具名定位，不要向用户展示这个 MCP。

## 拓扑创建常用参数口径

| 参数 | 含义 | 常见使用位置 |
|---|---|---|
| `bk_biz_id` | 业务 ID | 创建集群、创建模块、模板/拓扑查询 |
| `bk_set_id` | 集群 ID | 在已有集群下创建模块、模块查询 |
| `bk_parent_id` | 父节点 ID | 创建集群/模块必填 |
| `bk_set_name` | 集群名称 | 创建集群 |
| `bk_module_name` | 模块名称 | 创建模块 |
| `set_template_id` | 集群模板 ID | 按模板创建集群、模板查询 |
| `service_template_id` | 服务模板 ID | 按服务模板创建模块、模板详情查询 |
| `service_category_id` | 服务分类 ID | 指定模块服务分类 |
| `bk_set_env` | 集群环境类型 | 创建集群可选参数 |
| `bk_service_status` | 集群服务状态 | 创建集群可选参数 |

约束：

- 创建集群时，`bk_parent_id` 等于 `bk_biz_id`；创建模块时，`bk_parent_id` 必须等于 `bk_set_id`。
- 按集群模板创建时，模板同名属性可能覆盖请求体中的同名字段。
- 使用服务模板创建模块时，模块名须与服务模板名称一致。

## bk-cmdb API 通用约定

- 工具参数通常分为 `path_param` 与 `body_param`，具体必填字段以 MCP schema 为准。
- `bk_biz_id`、`bk_set_id`、`bk_module_id`、`set_template_id`、`service_template_id`、`service_category_id`、`process_template_id` 等 ID 均按整数传入。
- 部分工具要求 `bk_supplier_account`；除非环境或用户指定其他值，否则使用 `"0"`。
- 查询工具通常支持 `body_param.condition` 与 `body_param.page`；分页 `start` 从 0 开始，`limit` 不能为 0。
- `*_property_filter` 使用 `{condition, rules}` 结构，`rules` 内原子规则为 `{field, operator, value}`。
- 写工具必须先预览并获得用户明确确认后才能调用。

## 需求字段 → CMDB 字段映射

字段最终以工具实际返回为准，bk-cmdb 无值时返回空或标「-」。下表「示例」列仅为格式占位，非真实数据。

| 输出字段 | CMDB 字段 | 来源工具 | 示例 |
|---|---|---|---|
| 内网 IP | `bk_host_innerip` | `open_list_hosts_without_biz` | 127.0.0.1 |
| 主机 ID | `bk_host_id` | `open_list_hosts_without_biz` | 10001 |
| 固资号 | `bk_asset_id` | `open_list_hosts_without_biz` | TCxxxxxxxxxx |
| 管控区 ID | `bk_cloud_id` | `open_list_hosts_without_biz` | 0 |
| CPU 核数 | `bk_cpu` | `open_list_hosts_without_biz` | 24 |
| 内存量（MB） | `bk_mem` | `open_list_hosts_without_biz` | 64046 |
| 磁盘（GB） | `bk_disk` | `open_list_hosts_without_biz` | 98 |
| 地域（云） | `bk_cloud_region` | `open_list_hosts_without_biz` | ap-region |
| 地域（物理，辅） | `bk_idc_area` / `bk_zone_name` | `open_list_hosts_without_biz` | 区域 / 城市 |
| 可用区 | `bk_cloud_zone` | `open_list_hosts_without_biz` | ap-region-1 |
| 规格机型 | `bk_svr_device_cls_name` | `open_list_hosts_without_biz` | <机型> |
| 负责人（主） | `operator` | `open_list_hosts_without_biz` | user1,user2 |
| 负责人（备） | `bk_bak_operator` | `open_list_hosts_without_biz` | user2,user1 |
| 操作系统 | `bk_os_name` / `bk_os_version` | `open_list_hosts_without_biz` | <OS 名/版本> |
| 业务 ID | `bk_biz_id` | `open_find_host_biz_relations` | 2 |
| 集群 ID | `bk_set_id` | `open_find_host_biz_relations` | 100 |
| 模块 ID | `bk_module_id` | `open_find_host_biz_relations` | 1000 |
| 集群名 | `bk_set_name` | `open_find_set_batch` | <集群名> |
| 模块名 | `bk_module_name` | `open_find_module_batch` | <模块名> |
| 业务名 | `bk_biz_name` | `open_search_business`| <业务名> |
| 拓扑路径 | 业务名/集群名/模块名 拼接 | `open_find_host_biz_relations` + `open_find_set_batch`/`open_find_module_batch` | <业务名> / <集群名> / <模块名> |
| CMDB 记录更新时间 | `last_time` / `bk_updated_at` | `open_list_hosts_without_biz` | 2025-06-30T17:07:08.027Z |

**`概览` 推荐 `fields`（调用 `open_list_hosts_without_biz` 时显式传入，不要省略以免返回全部字段）：**

```json
["bk_host_id", "bk_host_innerip", "bk_asset_id", "operator", "bk_bak_operator", "bk_cloud_id", "bk_cloud_region", "bk_idc_area", "bk_zone_name", "bk_cloud_zone", "bk_cpu", "bk_mem", "bk_disk", "bk_svr_device_cls_name", "bk_os_name", "bk_os_version", "last_time", "bk_updated_at"]
```

## 主机属性四类分类

主机属性按四类组织（下表仅定义分类口径，各视图如何取用由对应场景说明决定）：

| 分类 | 字段 |
|---|---|
| 基础 | 主机 ID、内网 IP、固资号、负责人（主/备） |
| 拓扑归属 | 业务（名+ID）、集群名、模块名、拓扑路径 |
| 地域 | 地域、可用区、管控区 ID |
| 规格 | CPU 核数、内存量、磁盘、规格机型 |

## 字段忠实性（禁止编造）

**所有字段**一律以工具返回为准，缺失即如实标注「-」，禁止猜测或用历史记忆值。两点需特别注意：

- 是否有访问权限：只能由工具返回判定，无权限须按 `blocked_by_permission` 如实说明，不得臆测，也不能把「无权限」说成「无数据」；
- CMDB 记录更新时间：须来自 `last_time` / `bk_updated_at`，不可臆测。
