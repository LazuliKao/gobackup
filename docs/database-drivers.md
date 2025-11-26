# Additional Database Backup Drivers

GoBackup supports several additional MySQL/MariaDB backup drivers that provide different backup strategies and performance characteristics.

## mariadb-dump

MariaDB's native logical dump tool. Similar to `mysqldump` but specifically for MariaDB databases.

### Configuration

```yaml
databases:
  my_mariadb:
    type: mariadb-dump
    host: 127.0.0.1
    port: 3306
    database: my_database
    username: root
    password: secret
    # Optional: backup only specific tables
    tables:
      - users
      - orders
    # Optional: exclude specific tables
    exclude_tables:
      - logs
      - sessions
    # Optional: additional command-line arguments
    args: --single-transaction --quick
```

### Options

| Option           | Default     | Description                                   |
| ---------------- | ----------- | --------------------------------------------- |
| `type`           | -           | Must be `mariadb-dump`                        |
| `host`           | `127.0.0.1` | Database host                                 |
| `port`           | `3306`      | Database port                                 |
| `socket`         | -           | Unix socket path (overrides host/port if set) |
| `database`       | -           | **Required.** Database name to backup         |
| `username`       | `root`      | Database username                             |
| `password`       | -           | Database password                             |
| `tables`         | -           | List of specific tables to backup             |
| `exclude_tables` | -           | List of tables to exclude from backup         |
| `args`           | -           | Additional command-line arguments             |

---

## mysqlpump

MySQL's parallel dump utility. Offers better performance than mysqldump for large databases through parallel processing.

### Configuration

```yaml
databases:
  my_mysql:
    type: mysqlpump
    host: 127.0.0.1
    port: 3306
    database: my_database
    username: root
    password: secret
    # Number of parallel threads (default: 2)
    parallel: 4
    # Optional: backup only specific tables
    tables:
      - users
      - orders
    # Optional: exclude specific tables
    exclude_tables:
      - logs
      - sessions
    # Optional: additional command-line arguments
    args: --compress-output=ZLIB
```

### Options

| Option           | Default     | Description                                   |
| ---------------- | ----------- | --------------------------------------------- |
| `type`           | -           | Must be `mysqlpump`                           |
| `host`           | `127.0.0.1` | Database host                                 |
| `port`           | `3306`      | Database port                                 |
| `socket`         | -           | Unix socket path (overrides host/port if set) |
| `database`       | -           | **Required.** Database name to backup         |
| `username`       | `root`      | Database username                             |
| `password`       | -           | Database password                             |
| `parallel`       | `2`         | Number of parallel threads for dumping        |
| `tables`         | -           | List of specific tables to backup             |
| `exclude_tables` | -           | List of tables to exclude from backup         |
| `args`           | -           | Additional command-line arguments             |

---

## mydumper

High-performance multi-threaded backup tool for MySQL and MariaDB. Provides faster backups and restores compared to traditional dump tools.

### Configuration

```yaml
databases:
  my_mysql:
    type: mydumper
    host: 127.0.0.1
    port: 3306
    database: my_database
    username: root
    password: secret
    # Number of threads (default: 4)
    threads: 8
    # Enable compression (default: false)
    compress: true
    # Optional: backup only specific tables
    tables:
      - users
      - orders
    # Optional: additional command-line arguments
    args: --long-query-guard 300 --kill-long-queries
```

### Options

| Option     | Default     | Description                                   |
| ---------- | ----------- | --------------------------------------------- |
| `type`     | -           | Must be `mydumper`                            |
| `host`     | `127.0.0.1` | Database host                                 |
| `port`     | `3306`      | Database port                                 |
| `socket`   | -           | Unix socket path (overrides host/port if set) |
| `database` | -           | Database name to backup                       |
| `username` | `root`      | Database username                             |
| `password` | -           | Database password                             |
| `threads`  | `4`         | Number of threads for parallel dumping        |
| `compress` | `false`     | Enable compression of output files            |
| `tables`   | -           | List of specific tables to backup             |
| `args`     | -           | Additional command-line arguments             |

### Notes

- For table exclusion, use the `args` option with mydumper's `--regex` patterns
- mydumper creates multiple files in the output directory (one per table)
- Restoring requires using `myloader` command

---

## xtrabackup

Percona XtraBackup is a hot backup tool for MySQL/MariaDB that performs non-blocking, physical backups. Ideal for large databases where minimal downtime is critical.

### Configuration

```yaml
databases:
  my_mysql:
    type: xtrabackup
    host: 127.0.0.1
    port: 3306
    database: my_database
    username: root
    password: secret
    # Number of parallel threads (default: 1)
    parallel: 4
    # Enable compression (default: false)
    compress: true
    # Optional: additional command-line arguments
    args: --no-lock
```

### Options

| Option     | Default     | Description                                           |
| ---------- | ----------- | ----------------------------------------------------- |
| `type`     | -           | Must be `xtrabackup`                                  |
| `host`     | `127.0.0.1` | Database host                                         |
| `port`     | `3306`      | Database port                                         |
| `socket`   | -           | Unix socket path (overrides host/port if set)         |
| `database` | -           | Specific database to backup (backs up all if not set) |
| `username` | `root`      | Database username                                     |
| `password` | -           | Database password                                     |
| `parallel` | `1`         | Number of parallel threads for backup                 |
| `compress` | `false`     | Enable compression                                    |
| `args`     | -           | Additional command-line arguments                     |

### Notes

- XtraBackup performs physical (file-level) backups, not logical dumps
- Only available for x86_64 architecture
- Requires appropriate privileges (RELOAD, LOCK TABLES, PROCESS, REPLICATION CLIENT)
- Restoring requires using `xtrabackup --prepare` and `xtrabackup --copy-back` commands

---

## Comparison

| Feature             | mariadb-dump | mysqlpump  | mydumper  | xtrabackup                  |
| ------------------- | ------------ | ---------- | --------- | --------------------------- |
| Backup Type         | Logical      | Logical    | Logical   | Physical                    |
| Parallel Processing | No           | Yes        | Yes       | Yes                         |
| Hot Backup          | No           | No         | Yes       | Yes                         |
| Compression         | Via args     | Via args   | Built-in  | Built-in                    |
| Table Selection     | Yes          | Yes        | Yes       | Limited                     |
| Best For            | Small DBs    | Medium DBs | Large DBs | Large DBs, minimal downtime |

---

## Docker Support

When using the GoBackup Docker image, all these tools are pre-installed:

- `mariadb-dump` - Included with MariaDB client
- `mysqlpump` - MySQL-specific tool (may require MySQL client installation)
- `mydumper` - Compiled from source in the Docker image
- `xtrabackup` - Percona XtraBackup (x86_64 only)
