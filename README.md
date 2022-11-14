# Vector-controller

## Overview

[Vector](https://github.com/vectordotdev/vector) is a high-performance, end-to-end (agent & aggregator) observability data pipeline that puts you in control of your observability data.

vector-controller makes it's easier for Kubernetes user to use and config vector.

## Features

The vector controller includes but is not limited to, the following features:

* Specify the vector configuration with a Kubernetes Custom Resource, then it's will be merger into the target configmap that is a config file of a vector [agent](https://vector.dev/docs/setup/deployment/roles/#agent) or vector [aggregator](https://vector.dev/docs/setup/deployment/roles/#aggregator).

* Validate the config file of vector [sidecar](https://vector.dev/docs/setup/deployment/roles/#sidecar).