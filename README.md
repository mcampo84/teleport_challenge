---
authors: Matt Campo (matthew.f.campo@gmail.com)
state: draft
---

<!-- omit in toc -->
# Teleport Interview Challenge (Level 4)

<!-- omit in toc -->
## Contents
- [UX](#ux)
  - [start](#start)
  - [stop](#stop)
  - [status](#status)
  - [stream](#stream)
  - [cgroups](#cgroups)
    - [cgroup 1 - small jobs](#cgroup-1---small-jobs)
    - [cgroup 2 - medium jobs](#cgroup-2---medium-jobs)
    - [cgroup 3 - large jobs](#cgroup-3---large-jobs)
- [Process Execution Lifecycle (happy path)](#process-execution-lifecycle-happy-path)
- [API](#api)
- [Security Considerations](#security-considerations)
- [Test Plan](#test-plan)
- [Scope](#scope)
- [Design Approach](#design-approach)
- [Tradeoffs](#tradeoffs)

<!-- omit in toc -->
## Approvers
- [ ] @tigrato
- [ ] @russjones
- [ ] @fspmarshall
- [ ] @r0mant

<!-- omit in toc -->
## What
A secure remote command execution proxy with configurable cgroups for CPU, Memory and Disk I/O resource control.  
The proxy will be capable of supporting multiple clients concurrently, without deadlocking or leaking information.  
It will be run from the CLI of the client, be secured through mTLS and accessed via gRPC API, and provide `start`, `stop`, `status` and `stream` commands (details below).

<!-- omit in toc -->
## Why
From the original [challenge document](https://github.com/gravitational/careers/blob/main/challenges/systems/challenge-1.md#rationale):
> This exercise has two goals:
> 
> It helps us to understand what to expect from you as a developer, how you write production code, how you reason about API design and how you communicate when trying to understand a problem before you solve it.
> It helps you get a feel for what it would be like to work at Teleport, as this exercise aims to simulate our day-as-usual and expose you to the type of work we're doing here.
> We believe this technique is not only better, but also is more fun compared to whiteboard/quiz interviews so common in the industry. It's not without the downsides - it could take longer than traditional interviews.
> 
> Some of the best teams use coding challenges.
> 
> We appreciate your time and are looking forward to hack on this project together.

<!-- omit in toc -->
## Details
This project will consist of three distinct parts:
1. A reusable library, which contians the basic business logic for starting jobs on a server.
2. A gRPC API, which is the interface the clients will use to communicate with the server.
3. A CLI client that will communicate with the gRPC API.

### UX
The CLI will enable the client to execute the following, assuming they are authenticated and have all necessary permissions to execute the commands on the server:
1. [start](#start)
2. [stop](#stop)
3. [status](#status)
4. [stream](#stream)

#### start
The `start` command will be responsible for kicking off a job on a remote server. Provided the client has permissions to execute the job, the command will return a PID to the user.

**Examples**:  

Successful execution
```bash
> start [command] [arguments...]
Starting [command] with [arguments...]...
Success! PID: [pid]
```

Error
```bash
> start [command] [arguments...]
Starting [command] with [arguments...]...
Error: [error message]
```

#### stop
In order to stop a process, the user will need its PID and pass it to the `stop` command. Assuming the client has the necessary permissions, this will interrupt the execution of the job and return a success/failure message.

**Examples**:

Successful execution
```bash
> stop [pid]
Stopping [pid]...
Job [pid] successfully stopped!
```

Error
```bash
> stop [pid]
Stopping [pid]...
Error: [pid] could not be stopped. [error message]
```

#### status
In order to retrieve the status of a job, the user will needs its PID and pass it to the `status` command. Assuming the client has permission to query the status, and the job is active, this will return the job status, including CPU, memory, and disk I/O resource information.

**Examples**:

Successful execution
```bash
> status [pid]
Querying the status of [pid]...
Job [pid] is running.

=========================
|CPU:      | 23%        |
|Memory:   | 32MB       |
|Disk I/O: | 2.37k iops |
=========================
```

Error
```bash
> status [pid]
Querying the status of [pid]...
Error: Job [pid] not found
```

#### stream
In order to stream the output of a job, the user will need its PID and pass it to the `stream` command. Assuming the client has permission to strem the output, the output will be returned from the start of the process, and streamed to the CLI as new output is written.

**Examples**:

Successful execution
```bash
> stream [pid]
Starting output stream for [pid]...
================
[timestamp] [command] started
[timestamp] [log output]
[timestamp] [log output]
...
```

Error
```bash
> stream [pid]
Starting output stream for [pid]...
================
Error: [error message]
Exit [status]
```

#### cgroups
Since `cgroup`s are a unix feature, I won't be able to build and test the use of them on my local Mac. 
For this, I will need to either execute my code in a linux container, or a linux VM.
Since spinning up a linux container is easy enough, and I can bind/mount my development drive to that container for development use, that's how I'll build and test the features I'll be building.
I'll use a simple dockerfile to accomplish this. 

> **Note**: While tools like Docker Compose, Kubernetes, and Helm charts could greatly simplify the deployment and orchestration of this project, they are not included in this exercise due to time constraints. These tools would typically be used to manage containerized applications, ensuring scalability, reliability, and ease of deployment. However, for the purposes of this challenge, the focus will remain on the core functionality and security aspects of the project.

All resource allocations will be hard-coded for this exercise, and all allow-listed jobs will be mapped to a specific cgroup.

##### cgroup 1 - small jobs
* **CPU**: 10%
* **Memory**: 100MB
* **Disk I/O (bandwidth)**: 10MB/s

##### cgroup 2 - medium jobs
* **CPU**: 25%
* **Memory**: 1GB
* **Disk I/O (bandwidth)**: 50MB/s

##### cgroup 3 - large jobs
* **CPU**: 50%
* **Memory**: 8GB
* **Disk I/O (bandwidth)**: 100MB/s

### Process Execution Lifecycle (happy path)
1. Receive request
2. Client Authentication (middleware) - set client ID in context
3. Client Authorization
   1. For `start` ensure the client ID is permitted to execute the job they've requested
   2. For `stop`, `status` and `stream` commands, ensure the PID belongs to the client
4. `start`
   1. Using goroutine, start the requested process & capture the PID
      1. Begin capturing output in a buffer dedicated to the PID
      2. Assign the PID to a cgroup using a hard-coded mapping
      3. Map the PID to the client ID to ensure only this client may access the process
      4. Return the PID
5. `stop`
   1. Stop the job
   2. Close the job's output buffer
   3. Unset the PID in the ownership map
   4. Return a success/error message
6. `status`
   1. Return the job's status
7. `stream`
   1. Open the output buffer & stream all lines that exist
   2. Continue streaming the output until an interrupt signal is intercepted

### API
```proto
syntax = "proto3";

package teleport;

service CommandService {
    rpc Start (StartRequest) returns (StartResponse);
    rpc Stop (StopRequest) returns (StopResponse);
    rpc Status (StatusRequest) returns (StatusResponse);
    rpc StreamOutput (StreamOutputRequest) returns (stream StreamOutputResponse);
}

message StartRequest {
    string command = 1;
    repeated Argument arguments = 2;
}

message StartResponse {
    string pid = 1;
    string error = 2;
}

message StopRequest {
    string pid = 1;
}

message StopResponse {
    string result = 1;
    string error = 2;
}

message StatusRequest {
    string pid = 1;
}

message StatusResponse {
    string status = 1;
    string error = 2;
}

message StreamOutputRequest {
    string pid = 1;
}

message StreamOutputResponse {
    string output = 1;
    string error = 2;
}

message Argument {
    string name = 1;
    string value = 2;
}
```

### Security Considerations
1. While mTLS will be the method by which we authenticate the server and the client to one another, **permissions will be hard-coded for three separate clients for whom certificates will be generated for testing purposes**. The permissions will be a simple map of the client ID, associated with a discrete allow-list of commands that the client may execute.

2. No client will have visibility into the processes of another client. To achieve this, each PID will be stored in memory as it is running, and will map to its owner's client ID.

3. Allowing remote access to _all_ linux commands through this proxy sounds well beyond the scope of what this exercise seems to want me to do, and could be exploited by an attacker if permitted. 
Instead, an allow-list of commands will be provided. 
For this exercise, I'll hard-code it but normally that would be stored either in a config file or some other format that could then be deliberately edited by an authorized administrator.

<!-- omit from toc -->
#### TLS
Since TLS version 1.3 is the most recent version as of 2018, we'll use version 1.3. 

<!-- omit from toc -->
##### Server Certificate
1. For the purposes of this exercise, I will generate a certificate authority for a self-signed certificate using AES256 encryption.
```bash
# generate the CA private key using AES256
> openssl genpkey -algorithm RSA -out ca.key -aes256

# generate the CA certificate
> openssl req -new -x509 -key ca.key -sha256 -days 365 -out ca.crt
```
2. Generate the server key and CSR
```bash
# generate the server private key using AES256
> openssl genpkey -algorithm RSA -out server.key -aes256

# generate the server CSR
> openssl req -new -key server.key -out server.csr
```
3. Sign the server certificate with the CA
```bash
> openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256 
```
With these steps completed, we will have a complete set of credentials for the server to use for mTLS.

<!-- omit from toc -->
##### Client Certificate
We can reuse the same signing authority we set up for the server, since in a production environment, we would assume trust of a common certificate authority.
1. Generate the client key and CSR
```bash
# generate the client private key using AES256
> openssl genpkey -algorithm RSA -out client.key -aes256

# generate the client CSR
> openssl req -new -key client.key -out client.csr
```
2. Sign the client certificate with the CA
```bash
> openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256
```
With these steps completed, we will have a complete set of credentials for the client to use for mTLS.

### Test Plan
While I recognize the importance of unit testing and integration testing, given the time constraints for this exercise I will selectively test specific components.
Please refer to the [test plan](./doc/test_plan.md) for the behaviors I will evaluate.

### Scope
Since this is an exercise and not meant to be fully-scoped, we're going to stick to the meat of the exercise. That means the following will _not_ be considered:
* Remotely configurable `cgroup`s
* Deployment considerations / orchestration
* Containerization
* CI/CD - instead of running tests through github actions or another CI/CD pipeline, I'll run them locally and provide results in my PRs

### Design Approach
When designing this system, I'm considering the following:
1. The human user, who will be directly interacting with these commands
2. The client, which will need to authenticate and communicate with the server
3. The server, which will need to authenticate the client and execute all commands securely, according to business rules

I prefer to take a modular approach, defining and tesing the business logic for each component of the application in isolation before layering new functionality on top of it and testing the integration.

For example, some modules to be produced for the server:
1. ***Server*** - starts and listens for requests, forwarding them to the appropriate module for execution. Contains `AuthN` middleware for mTLS.
2. ***gRPC Service*** - handles the API requests. Calls `AuthZ` to allow/block commands from being executed by the `JobManager`.
3. ***AuthN*** - verifies the client certificate is valid, and maps the client certificate to a known client ID.
4. ***AuthZ*** - authorizes or denies the client access to execute and access jobs. Defines the mapping of client ID to allowed commands.
5. ***JobManager*** - starts and stops jobs, returns statuses, streams output. Defines the allow-list of commands, and assigns jobs to their appropriate cgroups.

For the client:
1. ***CLI Client*** - runs in the background, establishes a connection between the client and the server, listens for user input
2. ***AuthN*** - verifies the server certificate is valid

As each module is being developed, I'll start with the basic functionality (i.e. "`start` command kicks off a job and returns a PID"), and build upon it with subsequent PRs (i.e. "support starting/running concurrent processes"). 
This will allow me to build and test incrementally, in order to ensure quality and functionality. 

<!-- omit from toc -->
#### gRPC
Since I'll be using gRPC for the API, I will auto-generate the server and client code from the proto file.


### Tradeoffs
* Normally I would never leak certificates and keys by committing them to a repository, but in order to allow for proper evaluation I will include them in PRs as I prove out mTLS. In a real-world environment, I would provide one mocked set of each as test fixtures.
* I usually prefer to document designs using either UML or C4 models, but I didn't think it would be appropriate given the preference for basic markdown. 
Visual media like these tend to help convey a design better than text does, and helps me to better organize my thoughts when I can zoom in / out of a context, container or component.
* Given the guidelines of "4-5 pull requests" to build this feature, my PRs will likely be larger than I would typically make them. 
Normally I'd aim for no more than 100-200 lines of code per PR (not inclusive of tests).
