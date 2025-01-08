#!/bin/sh

# Create cgroups and set limits
mkdir -p /sys/fs/cgroup/small_jobs /sys/fs/cgroup/medium_jobs /sys/fs/cgroup/large_jobs

# Small jobs cgroup
echo 10000 > /sys/fs/cgroup/small_jobs/cpu.cfs_quota_us
echo 100000 > /sys/fs/cgroup/small_jobs/cpu.cfs_period_us
echo 104857600 > /sys/fs/cgroup/small_jobs/memory.limit_in_bytes
echo "10M" > /sys/fs/cgroup/small_jobs/blkio.throttle.write_bps_device

# Medium jobs cgroup
echo 25000 > /sys/fs/cgroup/medium_jobs/cpu.cfs_quota_us
echo 100000 > /sys/fs/cgroup/medium_jobs/cpu.cfs_period_us
echo 1073741824 > /sys/fs/cgroup/medium_jobs/memory.limit_in_bytes
echo "50M" > /sys/fs/cgroup/medium_jobs/blkio.throttle.write_bps_device

# Large jobs cgroup
echo 50000 > /sys/fs/cgroup/large_jobs/cpu.cfs_quota_us
echo 100000 > /sys/fs/cgroup/large_jobs/cpu.cfs_period_us
echo 8589934592 > /sys/fs/cgroup/large_jobs/memory.limit_in_bytes
echo "100M" > /sys/fs/cgroup/large_jobs/blkio.throttle.write_bps_device
