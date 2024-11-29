# Testing
The Go library right now has various tests defined. E.g. server interaction, oauth, discovery and signature verification tests.

To run the Go test suite, issue the following command in a shell

```bash
make test
```

Note that this runs the tests without any server interaction (so for now only the signature verification tests). To run the tests with an eduVPN server you need to specify environment variables:

```bash
SERVER_URI="eduvpn.example.com" PORTAL_USER="example" PORTAL_PASS="example" make test
```

This needs [python3-selenium](https://selenium-python.readthedocs.io/) and [geckodriver](https://github.com/mozilla/geckodriver/releases) (extract and put in your `$PATH`). Note that testing with a server assumes it uses a default portal, due to it needing to click on buttons on the web page. You can add your own portal by customizing the [called Selenium script](https://codeberg.org/eduVPN/eduvpn-common/src/branch/main/selenium_eduvpn.py).

There are other environment variables that can be used:

- `OAUTH_EXPIRED_TTL`: Use this for a server which has a low OAuth access token expiry time, e.g. 10 seconds. You would then set this variable to `"10"` so that a test is ran which waits for 10 seconds for the OAuth tokens to expire

## Testing the Python code
To test the Python code, issue the following command in a shell (you will need dependencies for all wrappers if you do this[^1]):

```bash
make -C wrappers/python test
```
