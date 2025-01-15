FEATURE cli/api interaction

SCENARIO command execution
    GIVEN the server has been authenticated
    WHEN the client executes a command
    THEN an API call is made to the server
