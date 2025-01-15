FEATURE configure cgroups to manage CPU, memory and disk I/O resources for jobs

SCENARIO a job is started
    GIVEN a job (command, PID)
    WHEN AssignToGroup is called
    THEN the job is assigned to a cgroup based on a preconfigured mapping
