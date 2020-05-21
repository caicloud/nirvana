# 项目结构和初始化

## 创建项目

Nirvana 创建项目非常简单，不过在创建项目之前，首先需要下载安装 Nirvana 的命令行工具：

```
$ go get -u github.com/caicloud/nirvana/cmd/nirvana
```

然后就可以直接使用命令创建项目（请确保 `$GOPATH/bin` 在 `$PATH` 中）：

```
$ cd $GOPATH/src/
$ nirvana init ./myproject
$ cd ./myproject
```

此时在 `$GOPATH/src/myproject` 会生成一个完整的 Nirvana 项目。项目结构如下：

```
.                                   #
├── .golangci.yml                   #
├── go.mod                          #
├── Makefile                        #
├── OWNERS                          #
├── README.md                       #
├── apis                            # Store apidocs (swagger json)
├── bin                             # Store the compiled binary
├── build                           # Store Dockerfile
│   └── demo-admin                  #
│       └── Dockerfile              #
├── cmd                             # Store startup commands for project
│   └── demo-admin                  #
│       └── main.go                 #
├── docs                            # Store docs
│   └── README.md                   #
├── hack                            # Store scripts
│   ├── README.md                   #
│   ├── read_cpus_available.sh      # Script to read available cpus
│   └── script.sh                   #
├── nirvana.yaml                    # File to describes your project
├── pkg                             # Store structures and converters required by API, distinguish by version
│   ├── apis                        #
│   │   ├── descriptors.go          # Store API descriptions (routing and others), distinguish by version
│   │   └── v1                      #
│   │       ├── converters          #
│   │       │   └── converters.go   #
│   │       ├── descriptors         #
│   │       │   ├── descriptors.go  #
│   │       │   └── message.go      # Store API definition of message
│   │       └── types.go            #
│   ├── filters                     # Store HTTP Request filter
│   │   └── filter.go               #
│   ├── handlers                    # Store the logical processing required by APIs
│   │   └── message.go              #
│   ├── middlewares                 # Store middlewares
│   │   └── middlewares.go          #
│   ├── modifiers                   # Store definition modifiers
│   │   └── modifiers.go            #
│   └── version                     # Store version information of project
│       └── version.go              #
├── test                            # Store all tests (except unit tests), e.g. integration, e2e tests.
│   └── test_make.sh                #
└── vendor                          #
```

这个项目中包含了编译和构建容器的基本工具（Makefile 和 Dockefile），还有 go mod 需要的包定义文件 `go.mod`。通过如下命令即可完成依赖包的安装：

```
$ go mod tidy
$ go mod vendor # 如果需要 vendor 的话
```

到这里就完成了整个项目的创建和依赖安装工作，默认的项目结构中自带了一个 API 范例，因此可以直接运行查看效果。

## 编译运行

### 直接编译运行

Nirvana 创建项目时自动生成了 Makefile，只需要使用简单的 `make` 命令就可以完成编译工作：

```
$ make build-local
```

在项目的 `bin` 目录中就能看到编译后的二进制文件，直接运行：

```
$ ./bin/myproject
```

启动时会打印出 Nirvana 的 Logo 以及当前项目的版本信息以及监听的端口，默认端口是 8080。

在服务启动之后，可以通过浏览器或者命令行访问 `http://localhost:8080/apis/v1/messages`：

```
$ curl http://localhost:8080/apis/v1/messages
```

就能够看到 API 的返回结果。

### 编译并打包成 Docker 镜像

在需要发布的时候，通常需要打包成镜像的形式，在 Makefile 中也提供了直接打包成镜像的命令：

```
$ make container
```

就会自动开始在容器内编译和打包镜像。不过这个过程中需要 `golang` 和 `debian:stretch` 这两个镜像。如果本地没有这两个镜像，或者希望使用其他镜像进行编译和构建工作，请修改 Makefile 和 `./build/myproject/Dockerfile` 或在使用 init 生成项目的时候使用 `--registry` 和 `--base-registry` 指定镜像仓库的地址。

打包完成后，可以通过 Docker 命令启动容器，然后进行访问：

```
$ docker run -p 8080:8080 myproject:v0.1.0
```

## Nirvana 项目配置

每个 Nirvana 项目都有一个 `nirvana.yaml` 配置文件，用于描述项目的基本信息和结构。

```yaml
# 项目名称
project: myproject
# 项目描述
description: This project uses nirvana as API framework
# 服务使用的协议，只能填写 http 和 https
schemes:
- http
# 访问 IP 或域名，可以有多个
hosts:
- localhost:8080
# 项目负责人
contacts:
- name: nobody
  email: nobody@nobody.io
  description: Maintain this project
# 项目 API 版本信息，用于区分不同版本的 API
# 用于文档和客户端生成
versions:
  # 版本名称
- name: v1
  # 版本描述
  description: The v1 version is the first version of this project
  # 版本规则
  rules:
    # 路径前缀，匹配前缀为 "/apis/v1" 的 API
  - prefix: /apis/v1
    # 正则表达式，用于匹配路径
    # 如果设置了 prefix，那么 regexp 字段无效
    regexp: ""
    # 这个字段仅用于在生成文档和客户端的时候，替换匹配的 API 路径。为空时不会进行替换。
    # 比如设置 replacement = "/apis/myproject/v1"
    # 那么 "/apis/v1/someapi" 为被替换为 "/apis/myproject/v1/someapi"
    replacement: ""
```

这个配置文件不会影响 Server 的运行，只用于描述项目的信息以及区分不同版本的 API。API 文档生成和客户端生成会依赖这个配置文件进行 API 版本识别和 API 路径替换，因此需要正确设置版本规则。
