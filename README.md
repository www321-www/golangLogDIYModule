# golangLogDIYModule
This package is designed to improve the weakness of golang builtin log module. In this package, you will be allowed to take Asynchronous logging and file content can be cut

1. Support output logs to different places
2. Log classification
  debug
  trace
  info
  warning
  error
  fatal
3. The log should support switch control. For example, it can be output at any level during development, but only the info level can be output after going online.
4. The complete log record should include time, line number, file name, log level, log information
5. Log files to be cut
  5.1 Cut by file size.
    Before each log is recorded, determine the size of the currently written file
  5.2 Cut by date.
    Set a field in the log structure to record the hours of the last cut
    Before writing the log, check whether the hours of the current time are consistent with those saved before. If they are inconsistent, they must     
    be cut.

Upgraded version: [function for writing files]
     The original code version is to write the log content (string) serially. If there are many logs to be written, it may cause the program to run slowly.
It is now required to change the synchronous log function into an asynchronous log write function
