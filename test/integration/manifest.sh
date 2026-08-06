podman pull docker-bkrepo.cwoa.net/h536b8/canway_d_arm/os/alpine:3.23.2-base
podman pull docker-bkrepo.cwoa.net/h536b8/canway_d/os/alpine:3.23.2-base
REPO=docker-bkrepo.cwoa.net/h536b8/canway_d/os/alpine:mult
podman manifest create $REPO
podman manifest add $REPO docker-bkrepo.cwoa.net/h536b8/canway_d_arm/os/alpine:3.23.2-base
podman manifest add $REPO docker-bkrepo.cwoa.net/h536b8/canway_d/os/alpine:3.23.2-base

#podman manifest push --all $REPO