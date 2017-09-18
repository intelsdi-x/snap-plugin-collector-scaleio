[![Build Status](https://travis-ci.org/intelsdi-x/snap-plugin-publisher-scaleio.svg?branch=master)](https://travis-ci.org/intelsdi-x/snap-plugin-publisher-scaleio)

# snap collector plugin - ScaleIO

This plugin supports collecting metrics from a ScaleIO cluster

It's used in the [Snap framework](http://github.com/intelsdi-x/snap).

1. [Getting Started](#getting-started)
   * [System Requirements](#system-requirements)
   * [Installation](#installation)
   * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
   * [Collected Metrics](#collected-metrics)
   * [Examples](#examples)
3. [Roadmap](#roadmap)
4. [Community Support](#community-support)
5. [Contributing](#contributing)
6. [License](#license)
7. [Acknowledgements](#acknowledgements)

## Getting Started

### System Requirements

* [golang 1.6+](https://golang.org/dl/) (needed only for building)

### Installation

#### Download File plugin binary:
You can get the pre-built binaries for your OS and architecture at the plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-collector-scaleio/releases) page.

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-collector-scaleio  
Clone repo into `$GOPATH/src/github.com/intelsdi-x/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-scaleio.git
```

Build the plugin by running make within the cloned repo:
```
$ make
```
This builds the plugin in `./build/`

### Configuration and Usage
First, be sure that you've familiarized yourself with the Snap framework by reading the
[Getting Started documentation](https://github.com/intelsdi-x/snap#getting-started).

This plugin requires several config options to run properly:

**Required**
* `gateway`: This is the URL of the gateway server.
* `username` and `password`: A username and password with the appropriate access to the REST API

**Optional**
* `verifySSL`: If set to `false` this disables SSL validation. This should not be used in production.

A full config example is below:

```
"/intel/scaleio": {
    "username": "admin",
    "password": "password",
    "gateway": "https://my-cluster",
    "verifySSL": false
}
```

## Documentation

### Collected Metrics
List of metrics collected by this plugin can be found in [METRICS.md file](METRICS.md).

### Examples
There is an example config found in the [examples directory](examples/file-collect.json).

All metrics are exposed with a dynamic namespace that encompasses each StoragePool. You can collect metrics from all of them or specify a storage pool that you are interested by putting its name instead wildcard - see how to specify the instance of dynamic metric in [Snap framework documentation](https://github.com/intelsdi-x/snap/blob/master/docs/TASKS.md#collect).

**Example watch**

```
NAMESPACE 							 DATA 		 TIMESTAMP
/intel/scaleio/96eb24f700000000/bckRebuildWriteBwc/numOccured 	 0 		 2016-07-08 23:16:02.304238351 -0700 PDT
/intel/scaleio/96eb24f700000000/pendingMovingOutBckRebuildJobs 	 0 		 2016-07-08 23:16:02.304238351 -0700 PDT
/intel/scaleio/96eb24f700000000/snapCapacityInUseInKb 		 3.145728e+06 	 2016-07-08 23:16:02.304238351 -0700 PDT
```

## Roadmap

This is currently in Alpha. Please let us know of any bugs you see.

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-scaleio/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-scaleio/pulls).

## Community Support
This repository is one of **many** plugins in **snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support) or join the Snap [Slack channel](https://intelsdi-x.herokuapp.com/).

## Contributing
We love contributions! 

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[Snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [Taylor Thomas](https://github.com/thomastaylor312)

And **thank you!** Your contribution, through code and participation, is incredibly important to us.
