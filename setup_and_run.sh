#!/bin/sh

# Create cgroups
sudo cgcreate -g cpu,memory,blkio:/small_jobs
sudo cgcreate -g cpu,memory,blkio:/medium_jobs
sudo cgcreate -g cpu,memory,blkio:/large_jobs

# Set limits for small jobs cgroup
sudo cgset -r cpu.cfs_quota_us=10000 small_jobs
sudo cgset -r cpu.cfs_period_us=100000 small_jobs
sudo cgset -r memory.limit_in_bytes=104857600 small_jobs
sudo cgset -r blkio.throttle.write_bps_device="8:0 10485760" small_jobs

# Set limits for medium jobs cgroup
sudo cgset -r cpu.cfs_quota_us=25000 medium_jobs
sudo cgset -r cpu.cfs_period_us=100000 medium_jobs
sudo cgset -r memory.limit_in_bytes=1073741824 medium_jobs
sudo cgset -r blkio.throttle.write_bps_device="8:0 52428800" medium_jobs

# Set limits for large jobs cgroup
sudo cgset -r cpu.cfs_quota_us=50000 large_jobs
sudo cgset -r cpu.cfs_period_us=100000 large_jobs
sudo cgset -r memory.limit_in_bytes=8589934592 large_jobs
sudo cgset -r blkio.throttle.write_bps_device="8:0 104857600" large_jobs

# Build the server application
go build -o server ./server/main.go

# Run the server application within the small_jobs cgroup
sudo cgexec -g cpu,memory,blkio:small_jobs ./server
