# bk-cmdb 主机查询 API 参考文档

本文档说明 bk-cmdb 主机查询相关工具的**使用场景、过滤条件与请求/返回示例**。**完整的参数定义（参数名/类型/必填）以各工具的 MCP schema 为准**，本文不再罗列参数表。

> 字段最终以接口实际返回为准，无值时返回空或标「-」。下文 IP、ID、名称均为占位示例。

### 参数结构

工具参数分为两类顶层键：

| 顶层键       | 说明                                          |
| ------------ | --------------------------------------------- |
| `body_param` | 请求体参数（查询条件、分页等嵌套其中）        |
| `path_param` | 路径参数（资源定位，如 `bk_biz_id`）          |

### 通用返回结构

所有工具返回统一的 JSON 信封，业务数据在 `data` 中：

```json
{
  "request_id": "uuid",
  "response_body": {
    "code": 0,
    "data": { ... },
    "message": "",
    "result": true
  },
  "status_code": 200
}
```

| 字段                    | 类型    | 说明                                          |
| ----------------------- | ------- | --------------------------------------------- |
| `response_body.code`    | integer | 业务状态码，`0` 表示成功，非 0 为错误         |
| `response_body.data`    | object  | 业务数据（各工具不同，见下方详细说明）        |
| `response_body.message` | string  | 错误或提示信息（如 `blocked_by_permission`）  |
| `response_body.result`  | boolean | 请求是否成功                                  |

### 通用嵌套类型

**主机/集群/模块属性过滤 `*_property_filter`：**

| 参数      | 类型   | 说明                                                                 |
| --------- | ------ | -------------------------------------------------------------------- |
| condition | string | 组合方式：`AND` / `OR`，最多嵌套 2 层                               |
| rules     | array  | 规则数组，每个元素是**原子规则** `{field, operator, value}`，或**嵌套组合** `{condition, rules}`（可再嵌套 AND/OR） |
| operator  | string | 原子规则的匹配方式：`equal`/`not_equal`/`in`/`not_in`/`less`/`less_or_equal`/`greater`/`greater_or_equal`/`between`/`not_between` |

`rules` 支持嵌套组合，可表达 `A AND (B OR C)` 这类带括号的复合条件（同一字段匹配多个值请用 `in`，`OR` 用于跨不同字段的「或」关系）：

```json
{
  "condition": "AND",
  "rules": [
    { "field": "bk_cloud_id", "operator": "equal", "value": 0 },
    {
      "condition": "OR",
      "rules": [
        { "field": "operator", "operator": "equal", "value": "user1" },
        { "field": "bk_bak_operator", "operator": "equal", "value": "user1" }
      ]
    }
  ]
}
```

**分页 page：**

| 参数  | 类型    | 说明                                       |
| ----- | ------- | ------------------------------------------ |
| start | integer | 起始位置，从 0 开始                        |
| limit | integer | 每页条数，不能为 0（不同工具上限不同）     |
| sort  | string  | 排序字段（部分工具支持）                   |

---

## 工具选型速查

先按「输入 → 输出」定位目标工具，再到下方对应小节看参数、权限与调用样例。

| 工具                            | 输入 → 输出                          | 所需权限 |
| ------------------------------- | ------------------------------------ | -------- |
| `open_list_hosts_without_biz`        | IP / 属性 → 主机属性 + `bk_host_id`  | 无       |
| `open_find_host_biz_relations`       | `bk_host_id` → 业务/集群/模块 ID     | 无       |
| `open_find_set_batch`                | `bk_set_id` → 集群名                 | 无       |
| `open_find_module_batch`             | `bk_module_id` → 模块名              | 无       |
| `open_search_business`               | 条件 → 业务列表（业务名）            | 无 |
| `open_list_biz_hosts`                | 业务（可按拓扑/属性过滤）→ 主机列表   | **业务访问**（`find_business_resource`） |
| `get_resource_pool_biz`              | 无入参 → 当前租户主机池 `bk_biz_id`  | **主机池主机查看** |

> 业务内查主机（含按集群/模块过滤、统计节点下主机数）统一用 `open_list_biz_hosts`：它支持按集群/模块（`bk_set_ids`/`bk_module_ids`）及属性过滤，并返回 `count` + 主机列表。

---

## 一、按主机标识符/属性查主机（入口，1 个工具）

### open_list_hosts_without_biz

不带业务 ID、按主机属性查询主机及其属性，返回含 `bk_host_id`。

**使用场景：** 已知主机标识符（内网 IP `bk_host_innerip` / 主机 ID `bk_host_id` / 固资号 `bk_asset_id`，按 `host_property_filter` 对应字段过滤即可，可 `OR` 组合或 `in` 批量），需要拿到主机的 `bk_host_id`、固资号、管控区、CPU/内存/磁盘、地域/可用区、机型、负责人、CMDB 更新时间等基础属性。是「按标识符定位主机」的首选。

过滤参数：`host_property_filter`（主机属性过滤，结构见「通用嵌套类型」）；`fields` 按视图取所需字段（见 `domain-terms.md`）。

**请求示例（单 IP）：**

```json
{
  "body_param": {
    "page": { "start": 0, "limit": 10 },
    "host_property_filter": {
      "condition": "AND",
      "rules": [
        { "field": "bk_host_innerip", "operator": "equal", "value": "127.0.0.1" }
      ]
    }
  }
}
```

**请求示例（多 IP 批量 + 精简字段）：**

```json
{
  "body_param": {
    "page": { "start": 0, "limit": 50 },
    "fields": ["bk_host_id", "bk_host_innerip", "bk_asset_id", "bk_cloud_id", "bk_cpu", "bk_mem", "bk_disk"],
    "host_property_filter": {
      "condition": "AND",
      "rules": [
        { "field": "bk_host_innerip", "operator": "in", "value": ["127.0.0.1", "127.0.0.2"] }
      ]
    }
  }
}
```

**请求示例（按主机 ID / 固资号过滤）：** 把过滤字段换成 `bk_host_id`（整数）或 `bk_asset_id`（字符串）即可，其余结构一致。固资号无固定格式，标识符不确定时用 `OR` 一次匹配多个候选字段（值为占位）：

```json
{
  "body_param": {
    "page": { "start": 0, "limit": 10 },
    "host_property_filter": {
      "condition": "OR",
      "rules": [
        { "field": "bk_host_id", "operator": "equal", "value": 10001 },
        { "field": "bk_asset_id", "operator": "equal", "value": "<asset_id>" }
      ]
    }
  }
}
```

**返回示例（`data`）：**

```json
{
  "count": 1,
  "info": [
    {
      "bk_host_id": 10001,
      "bk_host_innerip": "127.0.0.1",
      "bk_asset_id": "<asset_id>",
      "bk_cloud_id": 0,
      "bk_cpu": 24,
      "bk_mem": 64046,
      "bk_disk": 98,
      "bk_cloud_region": "ap-region",
      "bk_cloud_zone": "ap-region-1",
      "bk_svr_device_cls_name": "<机型>",
      "operator": "user1,user2",
      "bk_bak_operator": "user2,user1",
      "bk_os_name": "<OS>",
      "last_time": "2025-06-30T17:07:08.027Z"
    }
  ]
}
```

> `count` 为已找到主机总数，`0` 即未纳管 / not_found；`info` 元素字段以实际返回为准，中文含义见 `domain-terms.md`。

---

## 二、主机归属关系（host_id → 业务/集群/模块，1 个工具）

### open_find_host_biz_relations

按主机 ID 列表查询其业务/集群/模块归属关系。

**使用场景：** 已知 `bk_host_id`，需要解析主机挂在哪个业务（`bk_biz_id`）、集群（`bk_set_id`）、模块（`bk_module_id`）下，为后续解析名称、拼拓扑路径做准备。

**请求示例：**

```json
{
  "body_param": { "bk_host_id": [10001, 10002] }
}
```

**返回示例（`data`）：**

```json
[
  { "bk_host_id": 10001, "bk_biz_id": 2, "bk_set_id": 100, "bk_module_id": 1000 }
]
```

---

## 三、拓扑名称解析（ID → 名称，3 个工具）

### open_find_set_batch

按业务 ID + 集群 ID 列表批量查询集群详情。

**使用场景：** 已知 `bk_biz_id` 与 `bk_set_id`，批量解析集群名（`bk_set_name`），用于拼拓扑路径。

**请求示例：**

```json
{
  "path_param": { "bk_biz_id": 2 },
  "body_param": {
    "bk_ids": [100],
    "fields": ["bk_set_id", "bk_set_name"]
  }
}
```

**返回示例（`data`）：**

```json
[
  { "bk_set_id": 100, "bk_set_name": "<集群名>" }
]
```

---

### open_find_module_batch

按业务 ID + 模块 ID 列表批量查询模块详情。

**使用场景：** 已知 `bk_biz_id` 与 `bk_module_id`，批量解析模块名（`bk_module_name`），用于拼拓扑路径。

**请求示例：**

```json
{
  "path_param": { "bk_biz_id": 2 },
  "body_param": {
    "bk_ids": [1000],
    "fields": ["bk_module_id", "bk_module_name", "bk_set_id"]
  }
}
```

**返回示例（`data`）：**

```json
[
  { "bk_module_id": 1000, "bk_module_name": "<模块名>", "bk_set_id": 100 }
]
```

---

### open_search_business

**使用场景：** 已知 `bk_biz_id`，解析业务名（`bk_biz_name`）。

> 结果处理：
> - 命中：返回 `bk_biz_name`。
> - 未命中（业务名留空）：仅返回业务 ID，不可臆测。

过滤参数：`biz_property_filter`（业务属性过滤，结构见「通用嵌套类型」）。

**请求示例：**

```json
{
  "path_param": { "bk_supplier_account": "0" },
  "body_param": {
    "biz_property_filter": {
      "condition": "AND",
      "rules": [
        { "field": "bk_biz_id", "operator": "equal", "value": 2 }
      ]
    },
    "page": { "start": 0, "limit": 1 }
  }
}
```

**返回示例（`data`）：**

```json
{
  "count": 1,
  "info": [
    { "bk_biz_id": 2, "bk_biz_name": "<业务名>" }
  ]
}
```

> `count` 为匹配条件的业务数（无匹配时 `count=0`）。

---

## 四、按业务查主机（1 个工具）

### open_list_biz_hosts

查询业务下的主机列表（**不返回拓扑结构**），支持按集群/模块 ID（`bk_set_ids`/`bk_module_ids`）以及主机/集群/模块属性过滤。**需目标业务的 `find_business_resource`（业务访问）权限**，对无权限业务返回 `code 9900403 / 没有操作的权限`。

**使用场景：** 业务内查主机的统一接口。需要主机列表/属性时用它；要「按某个集群/模块过滤主机」或「统计某拓扑节点下主机数」时，用 `bk_set_ids`/`bk_module_ids`（或属性过滤）即可，返回的 `count` 即节点下主机数。

**参数说明：** 按主机属性过滤：`host_property_filter`（结构见「通用嵌套类型」）。

**请求示例（按模块过滤业务内主机）：**

```json
{
  "path_param": { "bk_biz_id": 2 },
  "body_param": {
    "bk_module_ids": [1000],
    "fields": ["bk_host_id", "bk_host_innerip"],
    "page": { "start": 0, "limit": 500 }
  }
}
```

**返回示例（`data`）：**

```json
{
  "count": 2,
  "info": [
    { "bk_host_id": 10001, "bk_host_innerip": "127.0.0.1" }
  ]
}
```

> `count` 为该业务内匹配条件的主机数（不含拓扑结构）；`info` 字段由 `fields` 决定，以实际返回为准。

---

## 二、当前租户主机池（internal MCP，1 个工具）

### get_resource_pool_biz

查询**调用方所在租户**的主机池业务 ID。该工具注册在单独的 internal MCP，不在对外 MCP 中，不要向用户暴露这个 MCP。无请求体，不要传租户 ID 或业务 ID。同一轮查询调用一次并复用结果。

**所需权限：** 主机池主机查看。

**返回示例（`data`）：**

```json
{
  "bk_biz_id": 1
}
```

`data.bk_biz_id` 即当前租户的主机池业务 ID。主机的 `bk_biz_id` 与它相等时，归属显示为「主机池」。
