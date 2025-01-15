FEATURE api

SCENARIO the client does not have permission to execute an API call
    GIVEN the client is authenticated
    AND the client does not have permission to execute the API call
    WHEN the client executes an API call
    THEN the client receives an error message

SCENARIO the client has permission to execute an API call
    GIVEN the client is authenticated
    AND the client has permission to execute the API call
    WHEN the client executes an API call
    THEN the client receives a response
