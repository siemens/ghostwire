# Based on https://www.redhat.com/sysadmin/podman-inside-container
ARG FEDORA_TAG

FROM fedora:${FEDORA_TAG}
RUN dnf -y install \
                procps systemd podman fuse-overlayfs \
                --exclude container-selinux && \
        dnf clean all && \
        rm -rf /var/cache /var/log/dnf* /var/log/yum.* && \
        systemctl enable podman.socket
RUN echo $'[containers]\n\
netns="host"\n\
userns="host"\n\
ipcns="host"\n\
utsns="host"\n\
cgroupns="host"\n\
cgroups="disabled"\n\
log_driver = "k8s-file"\n\
[engine]\n\
cgroup_manager = "cgroupfs"\n\
events_logger="file"\n\
runtime="crun"\n\
' > /etc/containers/containers.conf
CMD [ "/usr/sbin/init" ]
