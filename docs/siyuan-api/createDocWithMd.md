[← 返回首页](../index.html)

POST /api/filetree/createDocWithMd

# 通过Markdown创建文档

需要认证

## 接口描述

使用Markdown内容创建一个新的文档。此接口会将Markdown文本解析为思源笔记的块结构并创建文档。

## 请求参数

| 参数名 | 类型 | 必选 | 描述 |
|--------|------|------|------|
| notebook | string | 是 | 笔记本ID |
| path | string | 是 | 文档路径，不包含文件名，如"/foo/bar" |
| markdown | string | 是 | Markdown文本内容 |
| title | string | 否 | 文档标题，默认使用Markdown内容的第一个标题 |

## 返回值

| 参数名 | 类型 | 描述 |
|--------|------|------|
| code | number | 返回码，0表示成功 |
| msg | string | 返回信息 |
| data | object | 返回数据 |
| data.id | string | 新创建的文档ID |
| data.rootID | string | 文档根ID |
| data.box | string | 笔记本ID |
| data.path | string | 文档路径 |
| data.hPath | string | 人类可读的文档路径 |
| data.name | string | 文档名称 |

## 请求示例

### HTTP请求示例

```
POST /api/filetree/createDocWithMd HTTP/1.1
Host: 127.0.0.1:6806
Content-Type: application/json
Authorization: Token your-token

{
  "notebook": "20210808180117-czj9bvb",
  "path": "/教程",
  "markdown": "# Markdown示例\n\n这是一个通过API创建的Markdown文档。\n\n## 二级标题\n\n- 列表项1\n- 列表项2\n\n### 代码块\n\n```javascript\nconsole.log('Hello, World!');\n```",
  "title": "Markdown示例文档"
}
```

### 返回示例

```
{
  "code": 0,
  "msg": "",
  "data": {
    "id": "20230601120000-abcdef1",
    "rootID": "20230601120000-abcdef1",
    "box": "20210808180117-czj9bvb",
    "path": "/20210808180117-czj9bvb/20230601120000-abcdef1.sy",
    "hPath": "/教程/Markdown示例文档",
    "name": "Markdown示例文档"
  }
}
```

## Markdown支持的格式

```
# 一级标题
## 二级标题
### 三级标题

这是普通段落文本。

*斜体* **粗体** ***粗斜体***

> 这是引用文本

- 无序列表项1
- 无序列表项2
  - 嵌套列表项

1. 有序列表项1
2. 有序列表项2

[链接文本](https://example.com)

![图片描述](https://example.com/image.jpg)

`行内代码`

```javascript
// 代码块
function hello() {
  console.log('Hello World');
}
```

---

| 表格 | 表头 | 示例 |
|------|------|------|
| 内容1 | 内容2 | 内容3 |
| 行2  | 行2  | 行2  |
```

> 注意：这是一个社区维护的文档，可能与官方最新版本存在差异。
>
> 如果您觉得本文档有帮助，可以考虑赞助支持：[爱发电](https://afdian.com/a/leolee9086?tab=feed)

## 在线测试

API Token:

接口地址:

请求方法:

请求参数 (JSON):

发送请求

[返回首页](../index.html) [返回文档API](../pages/filetree.html) [上一个接口: 获取文档内容](getDoc.html) [下一个接口: 重命名文档](renameDoc.html)