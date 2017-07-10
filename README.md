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
* Add source information to Token, and use it in parse errors
* Handle types properly, don't just treat them as a dumb sequence of bytes.
  Test cases: long long, unsigned long, unsigned int, sequence<long, 10>,
  sequence<long,10>
* Handle identifiers properly. Foo::Bar is an identifier, struct Foo: Bar must
  not be treated as a struct of name "Foo:"

... more?
