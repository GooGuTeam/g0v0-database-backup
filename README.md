# MySQL 数据库备份服务

这是一个基于 Docker 和 Percona XtraBackup 的 MySQL 数据库备份服务。它支持全量备份、增量备份，并可以将备份上传到 Rclone 支持的云存储中。

## 环境搭建

### 配置文件

在启动服务之前，请确保项目根目录下存在以下配置文件：

- **config.json**: 服务配置文件

```json
{
    "mysqlUser": "backup",  // 备份用户
    "mysqlPassword": "your_password",  // 备份用户密码
    "mysqlHost": "mysql",  // MySQL 主机地址
    "mysqlPort": 3306,  // MySQL 端口
    "parallel": 4,  // 备份时的并行线程数
    "localBackupCount": 7,  // 本地保留的备份数量
    "rcloneRemote": "onedrive:"  // Rclone 远程名称
}
```

- **rclone.conf**: Rclone 配置文件，用于连接云存储。

### 启动服务

使用 Docker Compose 启动服务：

```bash
docker-compose up -d --build
```

服务启动后将监听 `32400` 端口。

## 2. API 使用说明

可以通过 HTTP 请求触发备份或下载任务。

### 触发全量备份

```bash
curl -X POST http://localhost:32400/full \
  -H "Content-Type: application/json" \
  -d '{"drive": "onedrive:", "comment": "This is a full backup"}'
```

### 触发增量备份

```bash
curl -X POST http://localhost:32400/incremental \
  -H "Content-Type: application/json" \
  -d '{"drive": "onedrive:", "comment": "This is an incremental backup"}'
```

### 下载备份

从云存储下载备份到本地的 `downloaded_backup` 目录。

```bash
curl -X POST http://localhost:32400/download \
  -H "Content-Type: application/json" \
  -d '{"drive": "onedrive:", "backup_name": "db_20251130_1200"}'
```

### 健康检查

```bash
curl http://localhost:32400/health
```

## 3. 备份恢复指南

本指南说明如何手动进入容器并使用 `xtrabackup` 恢复数据。

### 步骤 1: 进入容器

```bash
docker-compose exec backup-service bash
```

### 步骤 2: 定位备份文件

- 本地生成的备份位于 `backup` 目录。
- 从云端下载的备份位于 `downloaded_backup` 目录。

### 步骤 3: 解压备份

由于备份文件是使用 zstd 压缩的，在准备之前需要先解压。

```bash
# 假设备份目录为 /backup/db_20231027_1200
xtrabackup --decompress --target-dir=/backup/db_20231027_1200
```

### 步骤 4: 准备备份 (Prepare)

#### 情况 A: 仅恢复全量备份

```bash
xtrabackup --prepare --target-dir=/backup/db_20231027_1200
```

#### 情况 B: 恢复全量备份 + 增量备份

假设你有一个全量备份 `db_full` 和一个增量备份 `db_inc`。

1. 准备全量备份（注意使用 `--apply-log-only`）：

    ```bash
    xtrabackup --prepare --apply-log-only --target-dir=/backup/db_full
    ```

2. 将增量备份应用到全量备份上。如果有多个增量备份，可以依次应用。如果为最后一个增量备份，去掉 `--apply-log-only` 参数：

    ```bash
    xtrabackup --prepare --apply-log-only --target-dir=/backup/db_full --incremental-dir=/backup/db_inc
    ```

### 步骤 5: 恢复数据 (Copy-Back)

**警告**: 此操作将覆盖数据库的数据目录。请确保 MySQL 服务已停止，且数据目录为空（或已备份）。

1. 停止 MySQL 服务（在宿主机执行）：

    ```bash
    docker stop <mysql_container_name>
    ```

2. 清空 MySQL 数据目录（在 backup-service 容器内执行，请谨慎操作）：

    ```bash
    # 确保路径正确，例如：
    rm -rf /var/lib/mysql/*
    ```

3. 执行恢复（在 backup-service 容器内执行）：

    ```bash
    # 假设你要恢复到的目录是 /var/lib/mysql (容器内的挂载点)
    xtrabackup --copy-back --target-dir=/backup/db_20231027_1200 --datadir=/var/lib/mysql
    ```

4. 修复权限（在宿主机执行）：
    恢复后的文件可能属于 root 用户，需要修改为 mysql 用户。

    ```bash
    sudo chown -R 999:999 /path/to/mysql/data
    ```

5. 启动 MySQL 服务：

    ```bash
    docker start <mysql_container_name>
    ```
