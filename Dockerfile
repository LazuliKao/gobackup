# Stage 1: Build mydumper from source
FROM alpine:latest AS mydumper-builder
ARG MYDUMPER_VERSION="0.16.9-3"
RUN apk add --no-cache \
    curl \
    cmake \
    make \
    g++ \
    glib-dev \
    pcre-dev \
    zlib-dev \
    mariadb-dev
RUN cd /tmp && \
    curl -fLO "https://github.com/mydumper/mydumper/archive/refs/tags/v${MYDUMPER_VERSION}.tar.gz" && \
    tar xzf "v${MYDUMPER_VERSION}.tar.gz" && \
    cd "mydumper-${MYDUMPER_VERSION}" && \
    cmake . && \
    make && \
    make install

# Stage 2: Download and extract xtrabackup
FROM alpine:latest AS xtrabackup-downloader
ARG XTRABACKUP_VERSION="8.0.35-31"
RUN apk add --no-cache curl
RUN mkdir -p /xtrabackup-bin && \
    case "$(uname -m)" in \
      x86_64) \
        cd /tmp && \
        curl -fLO "https://downloads.percona.com/downloads/Percona-XtraBackup-8.0/Percona-XtraBackup-${XTRABACKUP_VERSION}/binary/tarball/percona-xtrabackup-${XTRABACKUP_VERSION}-Linux-x86_64.glibc2.17-minimal.tar.gz" && \
        tar xzf "percona-xtrabackup-${XTRABACKUP_VERSION}-Linux-x86_64.glibc2.17-minimal.tar.gz" && \
        cp percona-xtrabackup-${XTRABACKUP_VERSION}-Linux-x86_64.glibc2.17-minimal/bin/* /xtrabackup-bin/ ;; \
      *) echo 'XtraBackup not available for this architecture, skipping...' ;; \
    esac

# Stage 3: Download sqlpackage
FROM alpine:latest AS sqlpackage-downloader
RUN apk add --no-cache wget unzip
WORKDIR /tmp
RUN wget https://aka.ms/sqlpackage-linux && \
    unzip sqlpackage-linux -d /opt/sqlpackage && \
    chmod +x /opt/sqlpackage/sqlpackage

# Stage 4: Download influx CLI
FROM alpine:latest AS influx-downloader
ARG INFLUX_CLI_VERSION=2.7.5
RUN apk add --no-cache curl
RUN case "$(uname -m)" in \
      x86_64) arch=amd64 ;; \
      aarch64) arch=arm64 ;; \
      *) echo 'Unsupported architecture' && exit 1 ;; \
    esac && \
    curl -fLO "https://dl.influxdata.com/influxdb/releases/influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz" && \
    tar xzf "influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz" && \
    mkdir -p /influx-bin && \
    cp influx /influx-bin/influx

# Stage 5: Download etcdctl
FROM alpine:latest AS etcd-downloader
ARG ETCD_VER="v3.5.11"
RUN apk add --no-cache curl
RUN case "$(uname -m)" in \
      x86_64) arch=amd64 ;; \
      aarch64) arch=arm64 ;; \
      *) echo 'Unsupported architecture' && exit 1 ;; \
    esac && \
    curl -fLO https://github.com/etcd-io/etcd/releases/download/${ETCD_VER}/etcd-${ETCD_VER}-linux-${arch}.tar.gz && \
    tar xzf "etcd-${ETCD_VER}-linux-${arch}.tar.gz" && \
    mkdir -p /etcd-bin && \
    cp etcd-${ETCD_VER}-linux-${arch}/etcdctl /etcd-bin/etcdctl

# Stage 6: Final runtime image
FROM alpine:latest
ARG VERSION=latest

# Install runtime dependencies only
RUN apk add --no-cache \
  curl \
  ca-certificates \
  openssl \
  postgresql-client \
  mariadb-connector-c \
  mysql-client \
  mariadb-backup \
  redis \
  mongodb-tools \
  sqlite \
  # replace busybox utils
  tar \
  gzip \
  pigz \
  bzip2 \
  coreutils \
  # there is no pbzip2 yet
  lzip \
  xz-dev \
  lzop \
  xz \
  # pixz is in edge atm
  zstd \
  # 7z compression with password support
  7zip \
  # microsoft sql dependencies
  libstdc++ \
  gcompat \
  icu \
  # support change timezone
  tzdata \
  # mydumper runtime dependencies
  glib \
  pcre \
  zlib \
  mariadb-connector-c

# Copy mydumper from builder stage
COPY --from=mydumper-builder /usr/local/bin/mydumper /usr/local/bin/mydumper
COPY --from=mydumper-builder /usr/local/bin/myloader /usr/local/bin/myloader

# Copy xtrabackup binaries (if they exist)
COPY --from=xtrabackup-downloader /xtrabackup-bin/* /usr/local/bin/

# Copy sqlpackage
COPY --from=sqlpackage-downloader /opt/sqlpackage /opt/sqlpackage
ENV PATH="${PATH}:/opt/sqlpackage"

# Copy influx CLI
COPY --from=influx-downloader /influx-bin/influx /usr/local/bin/influx

# Copy etcdctl
COPY --from=etcd-downloader /etcd-bin/etcdctl /usr/local/bin/etcdctl

# Install gobackup
ADD install /install
RUN /install ${VERSION} && rm /install

CMD ["/usr/local/bin/gobackup", "run"]
