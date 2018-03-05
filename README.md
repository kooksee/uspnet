# gproject
> golang项目管理生成脚手架

## 项目初始化

### 请安装task项目管理工具

```shell
go get -u -v github.com/go-task/task/cmd/task
```

### 初始化项目

```shell
task init
```

### 安装依赖

```shell
task deps
```

    其他依赖请把依赖添加到`scripts/deps.sh中`

### Goland作为开发工具配置

    请把当前目录`pwd`配置为`Project GOPATH`

## 编译

```shell
task build
```
