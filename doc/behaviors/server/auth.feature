FEATURE server-side authentication/authorization

SCENARIO a client certificate is not authenticated
GIVEN an invalid client certificate
WHEN Authenticate is called
THEN false is returned

SCENARIO an authenticated client is not authorized to perform an action
GIVEN a client ID
AND the client is not authorized to perform the action
WHEN Authorize is called
THEN false is returned

SCENARIO an authenticated client is authorized to perform an action
GIVEN a client ID
AND the client is authorized to perform the action
WHEN Authorize is called
THEN true is returned
