# FoundationDB Database Adapter for GoBackup

This adapter provides backup support for FoundationDB using the `fdbbackup` command-line tool.

## Key Features

- **Backup Completion Guarantee**: Non-continuous backups wait until fully restorable before proceeding
- **One-time and Continuous Backups**: Support for both snapshot and continuous backup modes
- **Multiple Storage Backends**: Local filesystem and blob storage (S3, Azure, etc.)
- **Key Range Selection**: Backup specific key ranges
- **Distributed Backup**: Works with FoundationDB's distributed backup agents
- **Partitioned Logs**: Experimental support for partitioned mutation logs
- **Verification**: Automatically verifies backup directory exists after completion

## Prerequisites

- FoundationDB client tools installed (including `fdbbackup`)
- Running `backup_agent` processes on the cluster
- Cluster file accessible at the specified path

## Configuration Example

```yaml
models:
  my_app:
    databases:
      # Minimal configuration (recommended)
      # backup_url is optional - backups will be stored in GoBackup's dump directory
      # and processed by archive/compressor
      fdb_simple:
        type: foundationdb
        cluster_file: /etc/foundationdb/fdb.cluster
        
      # Basic local backup with explicit backup_url
      fdb_local:
        type: foundationdb
        cluster_file: /etc/foundationdb/fdb.cluster
        tag: daily
        backup_url: file:///backup/foundationdb
        
      # Continuous backup to S3
      fdb_continuous:
        type: foundationdb
        cluster_file: /etc/foundationdb/fdb.cluster
        tag: continuous
        continuous: true
        snapshot_interval: 3600
        backup_url: blobstore://@s3.amazonaws.com/fdb-backup?bucket=my-backups&region=us-west-2
        blob_credentials: /etc/foundationdb/blob_credentials.json
```

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `type` | - | Must be `foundationdb` |
| `cluster_file` | `/etc/foundationdb/fdb.cluster` | Path to cluster file |
| `tag` | `default` | Backup tag for managing multiple backups |
| `backup_url` | `file://<dump_path>` | **Optional.** Backup destination URL. If not specified, uses GoBackup's dump directory |
| `continuous` | `false` | Enable continuous backup mode |
| `snapshot_interval` | `864000` | Snapshot interval in seconds (10 days) |
| `partitioned_log` | `false` | Use partitioned logs (experimental) |
| `key_ranges` | - | List of key ranges to backup |
| `blob_credentials` | - | Path to blob credentials file |
| `args` | - | Additional fdbbackup arguments |

## Backup URL Formats

### Local Filesystem
```yaml
backup_url: file:///absolute/path/to/backup
```

### S3-Compatible Storage
```yaml
backup_url: blobstore://ACCESS_KEY:SECRET_KEY@s3.amazonaws.com/backup?bucket=my-bucket&region=us-west-2
```

### Azure Blob Storage
```yaml
backup_url: blobstore://ACCOUNT:KEY@account.blob.core.windows.net/backup?bucket=container
```

## Key Ranges

Specify key ranges to backup specific portions of your database:

```yaml
key_ranges:
  - "users "           # Keys starting with "users"
  - "orders products"  # Keys from "orders" to "products"
```

## Continuous vs One-time Backups

### One-time Backup
- Creates a point-in-time snapshot
- Backup completes and stops
- Use for scheduled full backups

### Continuous Backup
- Maintains near real-time backup
- Captures database mutations continuously
- Creates periodic snapshots
- Use for disaster recovery scenarios

## Restoring Backups

Use the `fdbrestore` command:

```bash
# Clear target key ranges first
fdbcli --exec "clearrange '' \xff"

# Start restore
fdbrestore start -r file:///backup/foundationdb \
  -t restore-tag \
  --dest-cluster-file /etc/foundationdb/fdb.cluster \
  -w
```

## Additional Resources

- [FoundationDB Backup Documentation](https://apple.github.io/foundationdb/backups.html)
- [GoBackup Database Drivers](../docs/database-drivers.md)

## Implementation Details

### Files

- `foundationdb.go` - Main adapter implementation
- `foundationdb_test.go` - Unit tests
- Integration in `base.go` - Registration of the adapter

### Commands Used

- `fdbbackup start` - Start a backup operation
- `fdbbackup status` - Check backup status
- `fdbbackup abort` - Abort running backup
- `fdbbackup discontinue` - Stop continuous backup

## Testing

Run the tests:

```bash
go test ./database -run TestFoundationDB -v
```

All tests should pass:
- TestFoundationDB_build
- TestFoundationDB_buildContinuous
- TestFoundationDB_buildWithKeyRanges
- TestFoundationDB_buildWithPartitionedLog
- TestFoundationDB_buildWithBlobStore
- TestFoundationDB_defaultBackupURL
