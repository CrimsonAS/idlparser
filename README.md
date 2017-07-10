# idlparser

Acts as a reader for the IDL file format, used in the DDS specification, for
example.

# status

It can read the DDS specification IDL, but there are a lot more things out
there that are not covered. Specifically:

* Type unions
* Struct inheritance
* ... probably more

It's also missing test coverage.

# TODO

* Tests
* Rename LexBuf to Lexer
* Rename ParseBuf to Parser
* Make ParseBuf take `[]Token` rather than a `*LexBuf`

... more?
