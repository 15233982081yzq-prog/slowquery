cat <<EOF >Dockerfile.alert

# BUILD_STAGE, ALWAYS use the golang image from bullseye/debian11 version
FROM golang:1.22-bullseye AS BUILD_STAGE

ENV ENV "test"

WORKDIR /data
RUN ln -sf /usr/share/zoneinfo/Asia/Singapore /etc/localtime
RUN git config --global url.https://gitlab-ci-token:${CI_JOB_TOKEN}@git.garena.com/.insteadOf https://git.garena.com/
RUN git clone https://git.garena.com/${CI_PROJECT_PATH}.git /data/smart-slowquery

WORKDIR /data/smart-slowquery
RUN git checkout ${CI_COMMIT_SHA}
RUN make build-alert

# PACK_STAGE
FROM --platform=linux/amd64 harbor.shopeemobile.com/rds-kube-operator/ubuntu-base:20.04.01 AS PACK_STAGE

# set commit id per https://jira.shopee.io/browse/SAI-4886
ARG COMMIT_ID=unknown
ENV COMMIT_ID=${CI_COMMIT_SHA}

# init workdir and working folder
WORKDIR /etc/slowquery/

# copy binary file and config file
COPY --from=BUILD_STAGE /data/smart-slowquery/bin/slowquery-alert /etc/slowquery/
RUN chmod +x /etc/slowquery/slowquery-alert

# change timezone to Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# run slowquery-openapi
ENTRYPOINT ["/etc/slowquery/slowquery-alert", "--config=/etc/slowquery/config/slowquery-alert.toml"]
EOF

cat <<EOF >Dockerfile.analyzer

# BUILD_STAGE, ALWAYS use the golang image from bullseye/debian11 version
FROM golang:1.22-bullseye AS BUILD_STAGE

ENV ENV "test"

WORKDIR /data
RUN ln -sf /usr/share/zoneinfo/Asia/Singapore /etc/localtime
RUN git config --global url.https://gitlab-ci-token:${CI_JOB_TOKEN}@git.garena.com/.insteadOf https://git.garena.com/
RUN git clone https://git.garena.com/${CI_PROJECT_PATH}.git /data/smart-slowquery

WORKDIR /data/smart-slowquery
RUN git checkout ${CI_COMMIT_SHA}
RUN make build-analyzer

# PACK_STAGE
FROM --platform=linux/amd64 harbor.shopeemobile.com/rds-kube-operator/ubuntu-base:20.04.01 AS PACK_STAGE

# set commit id per https://jira.shopee.io/browse/SAI-4886
ARG COMMIT_ID=unknown
ENV COMMIT_ID=${CI_COMMIT_SHA}

# init workdir and working folder
RUN mkdir -p /etc/slowquery/ && mkdir -p /tmp/log_analyzer/
WORKDIR /etc/slowquery/

# copy binary file and config file
COPY --from=BUILD_STAGE /data/smart-slowquery/bin/slowquery-analyzer /etc/slowquery/
RUN chmod +x /etc/slowquery/slowquery-analyzer

# change timezone to Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# run slowquery-analyzer
ENTRYPOINT ["/etc/slowquery/slowquery-analyzer", "--config=/etc/slowquery/config/slowquery-analyzer.toml"]
EOF

cat <<EOF >Dockerfile.cronjob

# BUILD_STAGE, ALWAYS use the golang image from bullseye/debian11 version
FROM golang:1.22-bullseye AS BUILD_STAGE

ENV ENV "test"

WORKDIR /data
RUN ln -sf /usr/share/zoneinfo/Asia/Singapore /etc/localtime
RUN git config --global url.https://gitlab-ci-token:${CI_JOB_TOKEN}@git.garena.com/.insteadOf https://git.garena.com/
RUN git clone https://git.garena.com/${CI_PROJECT_PATH}.git /data/smart-slowquery

WORKDIR /data/smart-slowquery
RUN git checkout ${CI_COMMIT_SHA}
RUN make build-cronjob

# PACK_STAGE
FROM --platform=linux/amd64 harbor.shopeemobile.com/rds-kube-operator/ubuntu-base:20.04.01 AS PACK_STAGE

# set commit id per https://jira.shopee.io/browse/SAI-4886
ARG COMMIT_ID=unknown
ENV COMMIT_ID=${CI_COMMIT_SHA}

# init workdir and working folder
RUN mkdir -p /etc/slowquery/ && mkdir -p /tmp/log_cronjob/
WORKDIR /etc/slowquery/

# copy binary file and config file
COPY --from=BUILD_STAGE /data/smart-slowquery/bin/slowquery-cronjob /etc/slowquery/
RUN chmod +x /etc/slowquery/slowquery-cronjob

# change timezone to Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# run slowquery-analyzer
ENTRYPOINT ["/etc/slowquery/slowquery-cronjob", "--config=/etc/slowquery/config/slowquery-cronjob.toml"]
EOF

cat <<EOF >Dockerfile.openapi

# BUILD_STAGE, ALWAYS use the golang image from bullseye/debian11 version
FROM golang:1.22-bullseye AS BUILD_STAGE

ENV ENV "test"

WORKDIR /data
RUN ln -sf /usr/share/zoneinfo/Asia/Singapore /etc/localtime
RUN git config --global url.https://gitlab-ci-token:${CI_JOB_TOKEN}@git.garena.com/.insteadOf https://git.garena.com/
RUN git clone https://git.garena.com/${CI_PROJECT_PATH}.git /data/smart-slowquery

WORKDIR /data/smart-slowquery
RUN git checkout ${CI_COMMIT_SHA}
RUN make build-openapi

# PACK_STAGE
FROM --platform=linux/amd64 harbor.shopeemobile.com/rds-kube-operator/ubuntu-base:20.04.01 AS PACK_STAGE

# set commit id per https://jira.shopee.io/browse/SAI-4886
ARG COMMIT_ID=unknown
ENV COMMIT_ID=${CI_COMMIT_SHA}

# init workdir and working folder
WORKDIR /etc/slowquery/

# copy binary file and config file
COPY --from=BUILD_STAGE /data/smart-slowquery/bin/slowquery-openapi /etc/slowquery/
RUN chmod +x /etc/slowquery/slowquery-openapi

# change timezone to Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# run slowquery-openapi
ENTRYPOINT ["/etc/slowquery/slowquery-openapi", "--config=/etc/slowquery/config/slowquery-openapi.toml"]
EOF

cat <<EOF >Dockerfile.platform

# BUILD_STAGE, ALWAYS use the golang image from bullseye/debian11 version
FROM golang:1.22-bullseye AS BUILD_STAGE

ENV ENV "test"

WORKDIR /data
RUN ln -sf /usr/share/zoneinfo/Asia/Singapore /etc/localtime
RUN git config --global url.https://gitlab-ci-token:${CI_JOB_TOKEN}@git.garena.com/.insteadOf https://git.garena.com/
RUN git clone https://git.garena.com/${CI_PROJECT_PATH}.git /data/smart-slowquery

WORKDIR /data/smart-slowquery
RUN git checkout ${CI_COMMIT_SHA}
RUN make build-platform

# PACK_STAGE
FROM --platform=linux/amd64 harbor.shopeemobile.com/rds-kube-operator/ubuntu-base:20.04.01 AS PACK_STAGE

# set commit id per https://jira.shopee.io/browse/SAI-4886
ARG COMMIT_ID=unknown
ENV COMMIT_ID=${CI_COMMIT_SHA}

# init workdir and working folder
WORKDIR /etc/slowquery/

# copy binary file and config file
COPY --from=BUILD_STAGE /data/smart-slowquery/bin/slowquery-platform /etc/slowquery/
RUN chmod +x /etc/slowquery/slowquery-platform

# change timezone to Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# run slowquery-platform
ENTRYPOINT ["/etc/slowquery/slowquery-platform", "--config=/etc/slowquery/config/slowquery-platform.toml"]
EOF

echo 'done'