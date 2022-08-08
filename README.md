# Go worker pool exercise

Go interview technical task about implementing a worker pool to process arbitrary jobs.

In this example, the task to perform is connecting to a list of domains, measuring average response time and data size of the index pages across all domains.
This is implemented in [main.go](main.go).

## Running

`go run .`

Available command-line flags are:

```
-input string
    input file name (default "input.csv")
-queue-size int
    size of job queue (default 20000)
-workers int
    number of workers to spawn (default 200)
```

Included in this repo is a sample input file with the first 20000 domains from Alexa top 1 million list.

## Improvements

This was developed in a very short time. Possible improvements to the library are:

- Testing and documentation
- Better stopping and cleanup of resources
- Job timeouts
- Logging
- Decoupling `Queue` and `Pool` to allow for multiple `Pool` implementations. (For example, allowing remote workers to connect to the pool)
- Applying generics to make result handling type-safe

## Original task statement

Your task is the design and implementation of a worker pool (whose size n can be configured)
that can be used for processing arbitrary tasks in a production environment.

As a use case you are required to take as input a list of domains (as a file with 1 domain per
line) and output the average response time and the average data size of the download of the
index pages across all of the input domains.

Since the list of domains can potentially be very large (or streamed) and unknown we want this
to be done in a controllable way. Doing it serially is too slow and doing everything at once is not
scalable.

You would use your generic work pool here to control the processing described above.

You do not have to save the index page anywhere. The output can just be to console. The
important bit is the design and implementation of the worker pool that will process the list of
domains and could be used for doing other tasks.

If you want to try a larger input data size please feel free to use the Alexa top 1 million list from
here and cut it down to whatever length you need -
https://s3.amazonaws.com/alexa-static/top-1m.csv.zip

If there is time you could add some unit tests but this is not expected.

Note: Other use cases of the generic worker pool we have as suggestions are implementing the
find or grep function on a large filesystem or sending emails. We don’t expect implementations
of any of these, they are just listed here to help you design the worker pool.