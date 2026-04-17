# gig view

查看当前项目的 `.gitignore` 文件内容。从当前目录向上查找，直到 Git 根目录。

## 用法

```
gig view [flags]
```

## 示例

```bash
# 查看当前项目的 .gitignore
gig view
```

## 注意

- 查找顺序：当前目录 → 上级目录 → ... → Git 根目录
- 如果整个项目中不存在 `.gitignore` 文件，会报错
- 会显示找到的文件完整路径
