# Docker 部署指南

本项目包含 docker-compose 配置，用于快速部署 XXL-JOB 服务器环境。

## 快速开始

启动 MySQL 和 XXL-JOB Admin：

```bash
docker-compose up -d
```

这将启动：
- **MySQL 8.0**：数据库服务，已自动初始化 xxl_job 数据库
  - 地址：localhost:3306
  - 用户名：xxl_job
  - 密码：xxl_job_123
  - Root 密码：root123456

- **XXL-JOB Admin**：管理后台
  - 地址：http://localhost:8080/xxl-job-admin
  - 默认用户：admin
  - 默认密码：123456

## 常用命令

```bash
# 查看运行中的容器
docker-compose ps

# 查看日志
docker-compose logs -f xxl-job-admin
docker-compose logs -f mysql

# 停止服务
docker-compose down

# 删除数据（包括数据库）
docker-compose down -v
```

## 访问地址

- XXL-JOB Admin：http://localhost:8080/xxl-job-admin
- MySQL：localhost:3306

## 数据库初始化

数据库初始化脚本位于 `xxl-job-admin/init.sql`，docker-compose 会在启动时自动执行。

如需手动初始化数据库：
```bash
docker exec -i xxl-job-mysql mysql -uroot -proot123456 < xxl-job-admin/init.sql
```

## 问题排查

### 连接超时
确保 MySQL 容器已完全启动。docker-compose 中设置了 healthcheck，会等待 MySQL 启动完成。

### 无法连接数据库
检查密码和用户名是否正确匹配，确保网络配置正确（使用 docker 网络）。

### 端口被占用
修改 docker-compose.yml 中的端口映射，例如：
```yaml
ports:
  - "3307:3306"  # 将宿主机的 3307 端口映射到容器的 3306 端口
```

## 参考资源

- [XXL-JOB 官方文档](https://www.xuxueli.com/xxl-job/)
- [XXL-JOB GitHub](https://github.com/xuxueli/xxl-job)
