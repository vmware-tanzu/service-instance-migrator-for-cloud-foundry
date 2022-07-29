# NOTICE

This package was copied from the bosh-cli [here][director] and modified to remove dependencies on 
[errors](https://github.com/cloudfoundry/bosh-utils/errors) and 
[logger](https://github.com/cloudfoundry/bosh-utils/logger)

It was chosen to copy this code in order to configure the socks5 proxy settings for tunneling without 
requiring the user of our software to set the `BOSH_ALL_PROXY` environment variable.

[director]: https://github.com/cloudfoundry/bosh-cli/tree/main/director
