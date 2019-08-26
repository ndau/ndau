# Ndau API SDK

Right now, it's easier for go code to interact directly with the blockchain
than to indirect through the REST API. That's because the `tool` package is, in
effect, a lean SDK.

However, it is desirable for as many services as possible to indirect through
the REST API<sup>[Citation Needed]</sup>. To make this practical, it needs to
be as easy as possible to use the API.

To that end, this package is designed to have largely the same interface that
the `tool` package does. It's not a drop-in replacement, but conversion should
be simple.
