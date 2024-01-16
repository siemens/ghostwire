# Based on https://www.redhat.com/sysadmin/podman-inside-container
ARG FEDORA_TAG

FROM fedora:${FEDORA_TAG}
RUN dnf -y install \
                procps systemd podman fuse-overlayfs \
                --exclude container-selinux && \
        dnf clean all && \
        rm -rf /var/cache /var/log/dnf* /var/log/yum.* && \
        systemctl enable podman.socket
CMD [ "/usr/sbin/init" ]
