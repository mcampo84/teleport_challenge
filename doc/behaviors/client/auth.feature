FEATURE client-side authentication

SCENARIO the server certificate is not authenticated
GIVEN the client is configured to authenticate the server
AND the server certificate is not authenticated
WHEN the client connects to the server
THEN the connection is refused

SCENARIO the server certificate is authenticated
GIVEN the client is configured to authenticate the server
AND the server certificate is authentic
WHEN the client connects to the server
THEN the connection is established
