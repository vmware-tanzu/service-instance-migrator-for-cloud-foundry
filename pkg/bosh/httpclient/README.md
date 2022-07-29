# NOTICE

This package was copied from the bosh-utils [here][httpclient] and [here][socks5] and modified to remove dependencies on
[errors](https://github.com/cloudfoundry/bosh-utils/errors) and
[logger](https://github.com/cloudfoundry/bosh-utils/logger). 

It was chosen to copy this code in order to configure the socks5 proxy settings for tunneling without 
requiring the user of our software to set the `BOSH_ALL_PROXY` environment variable.

[httpclient]: https://github.com/cloudfoundry/bosh-utils/blob/9f99bbf15b687cf78753188b4f93a698d814fade/httpclient/default_http_clients.go
[socks5]: https://github.com/cloudfoundry/bosh-utils/blob/24d8c72563a7f3a6d9c2bb4f7ce61de525954e23/httpclient/socksify.go