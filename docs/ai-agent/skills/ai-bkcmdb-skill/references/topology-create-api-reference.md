# bk-cmdb 集群拓扑创建 API 参考

本文记录 bk-cmdb 中与集群拓扑创建相关的常用 `open_` 工具及参数结构。实际以 MCP descriptor 为准。

## 通用约定

- 通用参数、ID 类型、分页与过滤结构见 `domain-terms.md`；完整参数定义以 MCP descriptor 为准。
- 写工具必须在预览并获得用户明确确认后才能调用。
- 大部分接口需要提供业务 ID；用户未提供时需先澄清。
- 通用返回信封、`page` 分页、`*_property_filter` 过滤结构等公共定义见 `host-query-api-reference.md`，本文不再重复。
- 查询/解析类工具 `open_search_business`、`open_find_set_batch`、`open_find_module_batch` 的完整参数与示例以 `host-query-api-reference.md` 为准，本文仅说明在创建链路中的用法。

## 解析业务

用户提供业务名而不是 `bk_biz_id` 时，使用 `open_search_business` 查询（参数结构与示例见 `host-query-api-reference.md` 的 `open_search_business`；此场景按 `bk_biz_name` 过滤）。

如果匹配到多个业务，要求用户选择。除非用户明确指定，否则不要按列表位置自行选择。

## open_search_set

查询业务下已有集群。

```json
{
  "path_param": {
    "bk_supplier_account": "0",
    "bk_biz_id": 123
  },
  "body_param": {
    "condition": {
      "bk_set_name": "cluster-name"
    },
    "page": {
      "start": 0,
      "limit": 20
    }
  }
}
```

## open_search_module

查询集群下已有模块。

```json
{
  "path_param": {
    "bk_supplier_account": "0",
    "bk_biz_id": 123,
    "bk_set_id": 456
  },
  "body_param": {
    "condition": {
      "bk_module_name": "module-name"
    },
    "page": {
      "start": 0,
      "limit": 20
    }
  }
}
```

## open_create_set

用户确认后，使用 `open_create_set` 创建集群。

### 关键规则

- 当指定 `set_template_id` 时，系统会按模板初始化集群属性，并自动创建模板关联模块。
- `bk_biz_id` 由 `path_param` 传入；`bk_set_name` 与 `bk_parent_id` 必须在 `body_param` 中传入。
- 创建集群时 `bk_parent_id` 必须等于 `bk_biz_id`；用户未给业务 ID 时先澄清，不发起写调用。

### 最小参数（直接创建）

```json
{
  "path_param": {
    "bk_biz_id": 123
  },
  "body_param": {
    "bk_set_name": "cluster-name",
    "bk_parent_id": 123
  }
}
```

### 按模板创建集群

```json
{
  "path_param": {
    "bk_biz_id": 123
  },
  "body_param": {
    "bk_set_name": "cluster-name",
    "bk_parent_id": 123,
    "set_template_id": 789
  }
}
```

### 错误处理

- 响应中包含 `bk_set_id` 时返回该 ID。若响应信息不足，先调用 `open_search_set` 或 `open_find_set_batch` 验证。
- 返回 `bk_set_name` 参数校验失败：通常是名称含非法字符或格式不合法；要求用户换名，不自动改名重试。

## open_create_module

用户确认后，使用 `open_create_module` 创建模块。

### 关键规则

- 创建模块时 `bk_parent_id` 必须等于 `bk_set_id`；缺少业务 ID 或集群 ID 时先澄清，不发起写调用。
- `service_category_id` 与 `service_template_id` 逻辑：
  - 两者都不传：使用默认服务分类。
  - 只传 `service_category_id`：使用指定服务分类。
  - 只传 `service_template_id`：从服务模板获取服务分类。
  - 两者都传：`service_category_id` 必须与 `service_template_id` 对应服务分类一致。
  - 使用服务模板创建模块时，模块名需与服务模板名称一致。

### 最小参数

```json
{
  "path_param": {
    "bk_biz_id": 123,
    "bk_set_id": 456
  },
  "body_param": {
    "bk_module_name": "module-name",
    "bk_parent_id": 456
  }
}
```

### 指定服务分类或服务模板

```json
{
  "path_param": {
    "bk_biz_id": 123,
    "bk_set_id": 456
  },
  "body_param": {
    "bk_module_name": "module-name",
    "bk_parent_id": 456,
    "service_category_id": 100,
    "operator": "admin",
    "bk_bak_operator": "backup_admin"
  }
}
```

批量创建模块时，按顺序逐个调用 `open_create_module` 并记录每个结果。

## 查询服务模板

创建模块前通常要查询服务模板，获取模板 ID、服务分类、进程模板等信息。

### open_list_service_template

查询业务下服务模板。

```json
{
  "body_param": {
    "bk_biz_id": 123,
    "page": {
      "start": 0,
      "limit": 20
    }
  }
}
```

若已知模板名，可通过 `search` + `is_exact` 过滤：

```json
{
  "body_param": {
    "bk_biz_id": 123,
    "search": "cmdb_coreserver",
    "is_exact": true,
    "page": {
      "start": 0,
      "limit": 20
    }
  }
}
```

### open_get_service_template

查询单个服务模板详情。

```json
{
  "path_param": {
    "service_template_id": 2000006270
  }
}
```

### open_list_proc_template

查询服务模板下进程模板。

```json
{
  "body_param": {
    "bk_biz_id": 123,
    "service_template_id": 2000006270,
    "page": {
      "start": 0,
      "limit": 20
    }
  }
}
```

### open_get_proc_template

查询单个进程模板详情。

```json
{
  "path_param": {
    "process_template_id": 2000004648
  },
  "body_param": {
    "bk_biz_id": 123
  }
}
```

## open_list_set_template

查询集群模板。

```json
{
  "path_param": {
    "bk_biz_id": 123
  },
  "body_param": {
    "page": {
      "start": 0,
      "limit": 20
    }
  }
}
```

使用 `open_list_set_template_related_service_template` 查询模板关联服务模板，用于预览模块结构。

```json
{
  "path_param": {
    "bk_biz_id": 123,
    "set_template_id": 789
  }
}
```

## open_sync_set_template_to_set

仅用于同步已有集群到模板最新状态，不是创建步骤。拓扑创建场景不得调用；按模板 `open_create_set` 时系统已自动创建模块。

## open_find_set_batch

批量查询业务下集群详情。验证创建结果时，可通过 `bk_ids` 指定集群ID，并在 `fields` 中加入 `create_time`、`set_template_id`，以确认是否新建成功。

```json
{
  "path_param": {
    "bk_biz_id": 123
  },
  "body_param": {
    "bk_ids": [456, 457],
    "fields": ["bk_set_id", "bk_set_name", "set_template_id", "create_time"]
  }
}
```

## open_find_module_batch

批量查询某业务的模块详情，验证创建结果时，可通过 `bk_ids` 指定模块ID，并在`fields` 中加入 `create_time`以确认是否新建成功。

```json
{
  "path_param": {
    "bk_biz_id": 123
  },
  "body_param": {
    "bk_ids": [111, 112],
    "fields": ["bk_module_id", "bk_module_name", "service_template_id", "create_time"]
  }
}
```

