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

## FoundationDB

FoundationDB's native backup tool. Supports continuous backups, distributed backup agents, and backup to local or blob storage (S3-compatible).

### Configuration

```yaml
databases:
  # Minimal configuration - backup_url is optional
  my_foundationdb:
    type: foundationdb
    cluster_file: /etc/foundationdb/fdb.cluster
    
  # Full configuration example
  my_foundationdb_full:
    type: foundationdb
    cluster_file: /etc/foundationdb/fdb.cluster
    tag: default
    backup_url: file:///backup/foundationdb  # Optional, defaults to GoBackup dump directory
    continuous: false
    snapshot_interval: 864000
    # Optional: specify key ranges to backup
    key_ranges:
      - "users "
      - "orders "
    # Optional: blob credentials file
    blob_credentials: /etc/foundationdb/blob_credentials.json
    # Optional: use partitioned log (experimental)
    partitioned_log: false
    # Optional: additional command-line arguments
    args: "--initial-snapshot-interval 60"
```

### Backup URL Formats

**Local Directory:**
```yaml
backup_url: file:///absolute/path/to/backup
```

**S3-compatible Blob Store:**
```yaml
backup_url: blobstore://ACCESS_KEY:SECRET_KEY@s3.amazonaws.com/backup-name?bucket=my-bucket&region=us-west-2
```

**Azure Blob Storage:**
```yaml
backup_url: blobstore://ACCOUNT_NAME:ACCESS_KEY@account.blob.core.windows.net/backup-name?bucket=container-name
```

### Options

| Option                | Default                         | Description                                                    |
| --------------------- | ------------------------------- | -------------------------------------------------------------- |
| `type`                | -                               | Must be `foundationdb`                                         |
| `cluster_file`        | `/etc/foundationdb/fdb.cluster` | Path to FoundationDB cluster file                              |
| `tag`                 | `default`                       | Backup tag name (for managing multiple backups)                |
| `backup_url`          | `file://<dump_path>`            | **Optional.** Backup destination URL (file:// or blobstore://) |
| `continuous`          | `false`                         | Enable continuous backup mode                                  |
| `snapshot_interval`   | `864000`                        | Snapshot interval in seconds (default: 10 days)                |
| `partitioned_log`     | `false`                         | Use partitioned logs (experimental, requires fast restore)     |
| `key_ranges`          | -                               | List of key ranges to backup (format: "BEGIN END")             |
| `blob_credentials`    | -                               | Path to blob credentials JSON file                             |
| `args`                | -                               | Additional command-line arguments for fdbbackup                |
| `before_script`       | -                               | Script to run before backup                                    |
| `after_script`        | -                               | Script to run after backup                                     |

### Key Range Format

Key ranges are specified as strings with optional begin and end:
- `"users "` - Backs up keys starting with "users" up to "users\xff"
- `"orders products"` - Backs up keys from "orders" to "products"
- `"apple banana"` and `"mango pineapple"` - Multiple ranges

### Continuous vs One-time Backup

**One-time Backup** (`continuous: false`):
- Creates a consistent point-in-time snapshot
- Backup completes and stops
- Use `-w` flag to wait for completion

**Continuous Backup** (`continuous: true`):
- Maintains near real-time backup
- Continuously captures database mutations
- Creates periodic snapshots based on `snapshot_interval`
- Backup runs indefinitely until stopped

### Blob Store URL Parameters

Common optional parameters for blob store URLs:

| Parameter                     | Short | Description                                       |
| ----------------------------- | ----- | ------------------------------------------------- |
| `secure_connection`           | `sc`  | Use HTTPS (1=yes, 0=no, default: 1)               |
| `max_send_bytes_per_second`   | `sbps`| Upload speed limit per agent                      |
| `max_recv_bytes_per_second`   | `rbps`| Download speed limit per agent                    |
| `concurrent_requests`         | `cr`  | Max concurrent requests                           |
| `request_timeout_min`         | `rtom`| Min request timeout in seconds                    |

Example with parameters:
```yaml
backup_url: blobstore://key:secret@s3.amazonaws.com/backup?bucket=my-backup&region=us-west-2&sbps=10485760
```

### Examples

**Minimal Configuration (Recommended for GoBackup Integration):**
```yaml
databases:
  fdb_simple:
    type: foundationdb
    cluster_file: /etc/foundationdb/fdb.cluster
```
This will backup to GoBackup's dump directory and the backup will be processed by archive/compressor.

**Basic Local Backup:**
```yaml
databases:
  fdb_local:
    type: foundationdb
    cluster_file: /etc/foundationdb/fdb.cluster
    backup_url: file:///backup/fdb/daily
```

**Continuous S3 Backup:**
```yaml
databases:
  fdb_continuous:
    type: foundationdb
    cluster_file: /etc/foundationdb/fdb.cluster
    tag: continuous
    continuous: true
    snapshot_interval: 3600  # 1 hour snapshots
    backup_url: blobstore://@s3.amazonaws.com/fdb-continuous?bucket=my-backups&region=us-west-2
    blob_credentials: /etc/foundationdb/blob_credentials.json
```

**Partial Backup (Specific Key Ranges):**
```yaml
databases:
  fdb_partial:
    type: foundationdb
    cluster_file: /etc/foundationdb/fdb.cluster
    tag: user-data
    backup_url: file:///backup/fdb/users
    key_ranges:
      - "user/ "
      - "profile/ "
```

### Notes

- FoundationDB backup requires `fdbbackup` command-line tool to be installed
- Backup agents (`backup_agent`) must be running on the cluster
- For blob storage, ensure all backup agents can access the storage endpoint
- The backup does not delete the target database - restore must be to a cleared database
- **Backup Completion**: Non-continuous backups use the `-w` flag, which blocks until the backup is complete and restorable
- Continuous backups run in the background and return immediately after starting
- Use `fdbbackup status -t <tag>` to check backup progress
- After backup completes, GoBackup verifies the backup directory exists before proceeding

### Restore

To restore a FoundationDB backup:

```bash
# Clear the target key ranges first
fdbcli --exec "clearrange '' \xff"

# Start restore
fdbrestore start -r file:///backup/fdb/daily -t restore-tag --dest-cluster-file /etc/foundationdb/fdb.cluster -w
```

See [FoundationDB Backup Documentation](https://apple.github.io/foundationdb/backups.html) for more details.

---

- `xtrabackup` - Percona XtraBackup (x86_64 only)
