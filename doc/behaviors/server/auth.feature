FEATURE server-side authentication/authorization

SCENARIO a client certificate is not authenticated
GIVEN the server is configured to authenticate the client
AND the client certificate is not authenticated
WHEN the client connects to the server
THEN the connection is refused

SCENARIO an authenticated client is not authorized to perform an action
GIVEN the client is authenticated
AND the client is not authorized to perform the action
WHEN the client performs the action
THEN the client receives an error message

SCENARIO an authenticated client is authorized to perform an action
GIVEN the client is authenticated
AND the client is authorized to perform the action
WHEN the client performs the action
THEN the client receives a response
