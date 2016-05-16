# SOPHOS UTM9 [![GoDoc](https://godoc.org/github.com/threez/sophos-utm9/confd?status.svg)](https://godoc.org/github.com/threez/sophos-utm9/confd)

All code written here is for private purposes. If you use the code in this
repository you might void your warranty. I can give no guarantees about it's
production readyness.

SOPHOS UTM9 can be downloaded here, for free (private use), as a software image:
[https://www.sophos.com/de-de/support/utm-downloads.aspx](https://www.sophos.com/de-de/support/utm-downloads.aspx)

## Confd

Simple client implementation, to access the configuration backend of the
SOPHOS UTM9.

## restd

restd is a little daemon that translates REST requests to the regular confd
daemon on sophos utm. The REST interface is exposed via a swagger spec and UI.

Authenticaion happens using basic auth. The same regular webadmin users can
be used. Same priviledges as in webadmin apply.

The server will not do https. The idea is, that it will be behind the httpd
running on the utm. httpd should do the ssl and handle proxing. This way, the
regular restictions (from where the webadmin is accessable) apply to restd too.

To inspect the API start the server and surf to
[http://localhost:3000/api/](http://localhost:3000/api/).

**THIS IS WIP AND NOT IN A USABLE STATE YET**

### TODO

* Integration into httpd
* cross compile && rpm
* Improve swagger documentation on the api interactions and model
* Allow error handling (add header for POST, PATCH, PUT, LOCK, UNLOCK)
  * actually return error
  * All (```X-RESTD-ERR-ACK: all```)
  * Last (```X-RESTD-ERR-ACK: last```)
  * None (```X-RESTD-ERR-ACK: none```)
* Implement PUT, PATCH, DELETE, LOCK, UNLOCK
  * Handle translation to confd object (for writes)
* Allow manipulation of nodes tree (GET, PUT)
* Create swagger API representation of nodes tree

### Ideas for restd

* Capture confd interactions (transactions) and create rest "methods" /
  "commands" as templates for later execution.
* Expose the reporting API
* Nagios integration
* Statsd integration
* Provisioning without SSH access (basic setup, ...)
* Monitoring and alerting functions
* WebSockets or SSI for firewall events / notifications
* Add http header for post requests (create object), that do insertion of
  that object into a node stucture, e.g.: ```X-RESTD-ADD-TO: /nodes/foo/bar```.
  The implementation should be using a transaction.
* Auditing API
* Allow Transactions similar to WebDAV transactions (with keepalive and header)

## License

MIT, see LICENSE file
