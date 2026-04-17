# gig version

显示 gig 的版本信息，包括版本号、Git 提交哈希和构建时间。

## 用法

```
gig version
```

## 示例

```bash
$ gig version
gig v1.0.0
commit: abcdef12
built:  2026-04-17T12:00:00
```

## 说明

- 版本号在编译时通过 `-ldflags` 注入
- Git 提交哈希显示前 8 位
- 未通过 ldflags 注入时显示 `dev`
