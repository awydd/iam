# iam
Identity and Access Management

## 环境

* **Go**: `1.27+`
* **Node.js (via NVM)**: `v26.7.0`
* **Package Manager**: `pnpm 11.24.0`
* **Database**: `MySQL 8.0.46`
* **Cache / Search**: `Redis Stack 7.4.0`

### 依赖工具

```bash
# 安装 Wire 依赖注入工具
go install github.com/google/wire/cmd/wire@latest
```

## 安装

```bash
git clone https://github.com/awydd/iam.git
# 或者使用 SSH 克隆
git clone git@github.com:awydd/iam.git
```

### 后端

```bash
cd iam

# 生成代码与依赖注入
go generate ./internal/infra/ent/generate.go
wire ./internal/wire/

go build
```

#### 初始化

首次运行需初始化**密钥对**与**管理员用户**：

```bash
./iam.exe init --username=eren --email=eren@awydd.com
```

> 💡 **提示**：系统管理员（`is_system`）权限仅支持通过命令行进行修改或维护：
> ```bash
> ./iam.exe user update-system --email <string> --password <string> --username <string>
> ```

### 前端

```bash
cd iam/web

# 配置文件
cp .example.env .env

# 安装依赖
pnpm i

# 启动
pnpm dev
```

完成上述步骤后，即可在浏览器中访问 [http://localhost:5173](http://localhost:5173)

## ⚖️ License

MIT License. See [LICENSE](./LICENSE).