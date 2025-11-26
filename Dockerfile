FROM alpine:latest
ARG VERSION=latest
RUN apk add \
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
  # microsoft sql dependencies \
  libstdc++ \
  gcompat \
  icu \
  # support change timezone
  tzdata \
  # mydumper build dependencies
  cmake \
  make \
  g++ \
  glib-dev \
  pcre-dev \
  zlib-dev \
  mariadb-dev \
  && \
  rm -rf /var/cache/apk/*

# Install mydumper from source
ARG MYDUMPER_VERSION="0.16.9-3"
RUN cd /tmp && \
    curl -fLO "https://github.com/mydumper/mydumper/archive/refs/tags/v${MYDUMPER_VERSION}.tar.gz" && \
    tar xzf "v${MYDUMPER_VERSION}.tar.gz" && \
    cd "mydumper-${MYDUMPER_VERSION}" && \
    cmake . && \
    make && \
    make install && \
    cd /tmp && \
    rm -rf "mydumper-${MYDUMPER_VERSION}" "v${MYDUMPER_VERSION}.tar.gz" && \
    mydumper --version

# Install Percona XtraBackup
# Note: XtraBackup is only available for x86_64 and requires glibc
ARG XTRABACKUP_VERSION="8.0.35-31"
RUN case "$(uname -m)" in \
      x86_64) \
        cd /tmp && \
        curl -fLO "https://downloads.percona.com/downloads/Percona-XtraBackup-8.0/Percona-XtraBackup-${XTRABACKUP_VERSION}/binary/tarball/percona-xtrabackup-${XTRABACKUP_VERSION}-Linux-x86_64.glibc2.17-minimal.tar.gz" && \
        tar xzf "percona-xtrabackup-${XTRABACKUP_VERSION}-Linux-x86_64.glibc2.17-minimal.tar.gz" && \
        cp percona-xtrabackup-${XTRABACKUP_VERSION}-Linux-x86_64.glibc2.17-minimal/bin/* /usr/local/bin/ && \
        rm -rf "percona-xtrabackup-${XTRABACKUP_VERSION}-Linux-x86_64.glibc2.17-minimal" \
               "percona-xtrabackup-${XTRABACKUP_VERSION}-Linux-x86_64.glibc2.17-minimal.tar.gz" && \
        xtrabackup --version ;; \
      *) echo 'XtraBackup not available for this architecture, skipping...' ;; \
    esac

WORKDIR /tmp
RUN wget https://aka.ms/sqlpackage-linux && \
    unzip sqlpackage-linux -d /opt/sqlpackage && \
    rm sqlpackage-linux && \
    chmod +x /opt/sqlpackage/sqlpackage

ENV PATH="${PATH}:/opt/sqlpackage"

# Install the influx CLI
ARG INFLUX_CLI_VERSION=2.7.5
RUN case "$(uname -m)" in \
      x86_64) arch=amd64 ;; \
      aarch64) arch=arm64 ;; \
      *) echo 'Unsupported architecture' && exit 1 ;; \
    esac && \
    curl -fLO "https://dl.influxdata.com/influxdb/releases/influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz" \
         -fLO "https://dl.influxdata.com/influxdb/releases/influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz.asc" && \
    tar xzf "influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz" && \
    cp influx /usr/local/bin/influx && \
    rm -rf "influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}" \
           "influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz" \
           "influxdb2-client-${INFLUX_CLI_VERSION}-linux-${arch}.tar.gz.asc" \
           "influx" && \
    influx version

# Install the etcdctl
ARG ETCD_VER="v3.5.11"
RUN case "$(uname -m)" in \
      x86_64) arch=amd64 ;; \
      aarch64) arch=arm64 ;; \
      *) echo 'Unsupported architecture' && exit 1 ;; \
    esac && \
    curl -fLO https://github.com/etcd-io/etcd/releases/download/${ETCD_VER}/etcd-${ETCD_VER}-linux-${arch}.tar.gz && \
    tar xzf "etcd-${ETCD_VER}-linux-${arch}.tar.gz" && \
    cp etcd-${ETCD_VER}-linux-${arch}/etcdctl /usr/local/bin/etcdctl && \
    rm -rf "etcd-${ETCD_VER}-linux-${arch}/etcdctl" \
           "etcd-${ETCD_VER}-linux-${arch}.tar.gz" && \
    etcdctl version



ADD install /install
RUN /install ${VERSION} && rm /install

CMD ["/usr/local/bin/gobackup", "run"]
