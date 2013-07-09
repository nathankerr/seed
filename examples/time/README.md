# Time Service

The time service returns the current time in a specified time zone.

# Running the server

    ./compile_and_run
    
will apply the network transformation to time.seed, compile the resulting go program (including the current_time_in map function), and start the server. A debug monitor will also be started on :8000.

# Running the client

In the client directory run

    go run client.go

then open :4000 in your web browser.
