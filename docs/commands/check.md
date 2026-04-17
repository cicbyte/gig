# gig check

检查指定文件是否被 `.gitignore` 忽略，显示所有匹配的规则及来源文件。

## 用法

```
gig check <file-path> [flags]
```

## 示例

```bash
# 检查文件是否被忽略
gig check .env

# 检查子目录中的文件
gig check dist/bundle.js
```

## 输出

如果文件被忽略，显示每个匹配规则的来源文件、行号和匹配模式，按优先级从高到低排列。

## 注意

- 优先级：子目录 `.gitignore` > 父目录 `.gitignore` > 根目录 `.gitignore`
- 如果文件未被忽略，会明确提示
