FEATURE start a job

SCENARIO the client does not have permisison to start a job
GIVEN the client is authenticated
AND the client does not have permission to start a job
WHEN the client starts a job
THEN the client receives an error message

SCENARIO the client has permission to start a job
GIVEN the client is authenticated
AND the client has permission to start a job
WHEN the client starts a job
THEN the job is started

FEATURE stop a job

SCENARIO the client does not have permission to stop a job
GIVEN the client is authenticated
AND the client does not have permission to stop a job
WHEN the client stops a job
THEN the client receives an error message

SCENARIO the job cannot be found
GIVEN the client is authenticated
AND the client has permission to stop a job
WHEN the client stops a job
THEN the client receives an error message

SCENARIO the client has permission to stop a job in progress
GIVEN the client is authenticated
AND the client has permission to stop a job
AND the job is in progress
WHEN the client stops the job
THEN the job is stopped

FEATURE fetch a job's status

SCENARIO the client does not have permission to query a job's status
GIVEN the client is authenticated
AND the client does not have permission to query a job's status
WHEN the client queries a job's status
THEN the client receives an error message

SCENARIO the job cannot be found
GIVEN the client is authenticated
AND a job ID
AND the job does not exist
WHEN the client queries a job's status
THEN the client receives an error message

SCENARIO the client has permission to query a job's status
GIVEN the client is authenticated
AND a job ID
AND the client has permission to query a job's status
WHEN the client queries a job's status
THEN the client receives the job's status

FEATURE stream a job's output

SCENARIO the client does not have permission to stream a job's output
GIVEN the client is authenticated
AND the client does not have permission to stream a job's output
WHEN the client streams a job's output
THEN the client receives an error message

SCENARIO the job cannot be found
GIVEN the client is authenticated
AND a job ID
AND the job does not exist
WHEN the client streams a job's output
THEN the client receives an error message

SCENARIO the client has permission to stream a job's output
GIVEN the client is authenticated
AND a job ID
AND the client has permission to stream a job's output
WHEN the client streams a job's output
THEN the client receives the job's output
